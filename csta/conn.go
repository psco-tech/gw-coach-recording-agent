package csta

import (
	"bufio"
	"context"
	"encoding/binary"
	"encoding/xml"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"reflect"
	"strconv"
	"sync"
	"time"
)

const defaultDialTimeout = 30 * time.Second
const defaultSessionDuration = 60

var defaultHandlers = map[MessageType]HandleFunc{
	MessageTypeSystemStatus: acknowledgeSystemStatus,
}

type ConnectionState int

const (
	ConnectionStateError           ConnectionState = -1
	ConnectionStateIdle            ConnectionState = 0
	ConnectionStateActive          ConnectionState = 1
	ConnectionStateStartingSession ConnectionState = 2
	ConnectionStateClosed          ConnectionState = 3
)

type Conn interface {
	Write(invokeId uint, m Message) error
	Read() (invokeId uint, m Message, err error)
	State() ConnectionState
	Request(request Message, responseHandler HandleFunc) error
	Close() error
	Closed() <-chan struct{}

	Handle(messageType MessageType, responseHandler HandleFunc)
	RemoveHandler(messageType MessageType)

	// Application Session Services
	StartApplicationSession(applicationId string, applicationSpecificInfo interface{}, protocolVersion string, callbacks ...HandleFunc) error

	// Monitoring Services
	MonitorStart(monitorObject CSTAObject, monitorType MonitorType, callback ...HandleFunc) error
}

type ConnectionOptions struct {
	DisableImmediateFlushing bool
}

type cstaConn struct {
	mutex        sync.Mutex
	ctx          context.Context
	lastInvokeId uint
	timeout      time.Duration
	options      *ConnectionOptions
	conn         net.Conn
	rw           *bufio.ReadWriter
	state        ConnectionState
	sessionId    string
	closed       context.CancelFunc

	handlers     map[MessageType]HandleFunc
	transactions map[uint]HandleFunc
}

type Context struct {
	conn     Conn
	InvokeID uint
	Message  Message
	Error    error
}

func (c *Context) error(err error) {
	// TODO send an error
}

type HandleFunc func(c *Context)

func (c *cstaConn) Handle(messageType MessageType, responseHandler HandleFunc) {
	c.handlers[messageType] = responseHandler
}

func (c *cstaConn) RemoveHandler(messageType MessageType) {
	delete(c.handlers, messageType)
}

// Write marshals and writes a CSTA message to the underlying connection
func (c *cstaConn) Write(invokeId uint, message Message) error {
	msg, err := marshal(invokeId, message)
	if err != nil {
		return fmt.Errorf("failed to marshal CSTA message: %w", err)
	}

	_, err = c.rw.Write(msg)
	if err != nil {
		return fmt.Errorf("failed to write CSTA message: %w", err)
	}

	if !c.options.DisableImmediateFlushing {
		c.rw.Flush()
	}

	return nil
}

// Read reads a complete CSTA message from the connection and unmarshals it
func (c *cstaConn) Read() (uint, Message, error) {
	// Read a CSTA header
	cstaHeader := make([]byte, 8)
	_, err := io.ReadFull(c.rw, cstaHeader)
	if err != nil {
		// Pass down EOF
		if errors.Is(err, io.EOF) {
			c.state = ConnectionStateClosed
			return 0, nil, err
		}
		c.state = ConnectionStateError
		return 0, nil, fmt.Errorf("failed to read CSTA header: %w", err)
	}

	// Get the message length from the CSTA header
	length := binary.BigEndian.Uint16(cstaHeader[2:4])

	// Quick sanity check for the message length
	if length <= 8 {
		return 0, nil, fmt.Errorf("invalid message length")
	}

	iid, err := strconv.ParseInt(string(cstaHeader[4:8]), 10, 0)
	if err != nil {
		return 0, nil, fmt.Errorf("failed to parse the invoke ID: %w", err)
	}
	invokeId := uint(iid)

	// Allocate a buffer for the message body
	body := make([]byte, length-8)

	// Read the message body from the wire
	_, err = io.ReadFull(c.rw, body)
	if err != nil {
		return invokeId, nil, fmt.Errorf("failed to read the message body: %w", err)
	}

	// Generic message to get the root element
	cstaMessage := struct {
		XMLName xml.Name
	}{}

	err = xml.Unmarshal(body, &cstaMessage)
	if err != nil {
		return invokeId, nil, fmt.Errorf("failed to get the message type: %w", err)
	}

	// Try to get an instance of the message type to unmarshal
	if messageType, ok := messageTypes[MessageType(cstaMessage.XMLName.Local)]; ok {
		instance := reflect.New(messageType).Interface()
		m := instance.(Message)

		// Unmarshal into message type
		err = unmarshal(append(cstaHeader, body...), m)
		if err != nil {
			return invokeId, nil, fmt.Errorf("failed to unmarshal message: %w", err)
		}

		return invokeId, m, nil
	}

	return invokeId, nil, fmt.Errorf("unhandled message type: %s", cstaMessage.XMLName.Local)
}

// Dial establishes a new connection to a switching function with default timeout paramters
func Dial(network string, address string, ctx context.Context, options *ConnectionOptions) (Conn, error) {
	return DialTimeout(network, address, defaultDialTimeout, ctx, options)
}

// DialTimeout establishes a new connection to a switching function
func DialTimeout(network string, address string, timeout time.Duration, ctx context.Context, options *ConnectionOptions) (Conn, error) {
	if options == nil {
		options = &ConnectionOptions{}
	}

	dialContext, cancelDialing := context.WithTimeout(ctx, timeout)
	defer cancelDialing()

	var cstaDialer net.Dialer

	// Establish a connection with the switching function
	tcpConn, err := cstaDialer.DialContext(dialContext, network, address)
	if err != nil {
		return nil, fmt.Errorf("failed to establish a connection with the switching function: %w", err)
	}

	ctx, closed := context.WithCancel(ctx)

	conn := cstaConn{
		timeout:      timeout,
		ctx:          ctx,
		closed:       closed,
		options:      options,
		conn:         tcpConn, // Keep a reference to the underlying net.Conn
		rw:           bufio.NewReadWriter(bufio.NewReader(tcpConn), bufio.NewWriter(tcpConn)),
		handlers:     defaultHandlers,
		transactions: make(map[uint]HandleFunc),
	}

	go conn.messageHandler()

	return &conn, nil
}

func (c *cstaConn) messageHandler() {
	for {
		invokeId, message, err := c.Read()
		if err != nil {
			if errors.Is(err, io.EOF) {
				log.Printf("PBX connection lost: %s\n", err)
				c.state = ConnectionStateClosed
				c.Close()
				return
			}
			c.state = ConnectionStateError
			log.Printf("Failed to Read() from CSTA connection: %s\n", err)
			continue
		}

		messageContext := &Context{
			conn:     c,
			Message:  message,
			InvokeID: invokeId,
		}

		// If there is a handler for this specific request, run it
		if tx, ok := c.transactions[invokeId]; ok {
			go tx(messageContext)
			delete(c.transactions, invokeId)

			continue
		}

		// If there is a handler for this message type, run it
		if handler, ok := c.handlers[message.Type()]; ok {
			go handler(messageContext)

			continue
		}

		log.Printf("Failed to find a handler for message of type: %s\n", message.Type())

		// We didn't find a handler for this specific message

		// TODO reply with error
	}
}

type Error struct {
}

type Response struct {
	Message *Message
	Error   error
}

func (c *cstaConn) Close() error {
	log.Printf("Closing CSTA connection\n")
	if c.conn != nil {
		c.conn.Close()
	}

	// Preserve the error state
	if c.state != ConnectionStateError {
		c.state = ConnectionStateClosed
	}

	// Notify listeners that we're closed
	c.closed()

	return nil
}

func (c *cstaConn) Request(request Message, responseHandler HandleFunc) error {
	// Get a free invoke ID
	requestId := c.nextInvokeID()

	// Register a handler for the response
	c.transactions[requestId] = responseHandler

	// Send out the request
	return c.Write(requestId, request)
}

func (c *cstaConn) State() ConnectionState {
	return c.state
}

func (c *cstaConn) nextInvokeID() uint {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.lastInvokeId += 1
	if c.lastInvokeId >= 9000 {
		c.lastInvokeId = 0
	}
	return c.lastInvokeId
}

func dispatchCallbacks(ctx *Context, callbacks ...HandleFunc) {
	for _, cb := range callbacks {
		cb(ctx)
	}
}

func (c *cstaConn) Closed() <-chan struct{} {
	return c.ctx.Done()
}
