package csta

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
)

const cstaHeaderSize = 8
const formatIndicatorTCPWithoutSOAP uint16 = 0x0000

type MessageType string

type Message interface {
	Type() MessageType
}

// Registry for all implemented message types, populate via registerMessageType(MessageType, reflect.Type)
var messageTypes map[MessageType]reflect.Type

func registerMessageType(messageType MessageType, implementation reflect.Type) {
	if messageTypes == nil {
		messageTypes = make(map[MessageType]reflect.Type)
	}

	messageTypes[messageType] = implementation
}

// Generic marshal implementation, most messages can just call this to marshal themselves
func marshal(invokeId uint, m Message) ([]byte, error) {
	body, err := xml.Marshal(m)
	if err != nil {
		return []byte{}, fmt.Errorf("failed to marshal CSTA message body: %w", err)
	}

	msg := new(bytes.Buffer)

	msg.Write([]byte{0x00, 0x00}) // TCP without SOAP format indicator
	binary.Write(msg, binary.BigEndian, uint16(len(body)+cstaHeaderSize))
	msg.WriteString(fmt.Sprintf("%04d", invokeId))
	msg.Write(body)

	return msg.Bytes(), nil
}

// Generic unmarshal implementation that verifies a message
// ignores the invoke id and unmarshals the body into a message implementation
func unmarshal(data []byte, m Message) error {
	msgReader := bufio.NewReader(bytes.NewBuffer(data))

	// Read the format indicator
	var formatIndicator uint16
	err := binary.Read(msgReader, binary.BigEndian, &formatIndicator)
	if err != nil {
		return fmt.Errorf("failed to read format indicator: %w", err)
	}

	// Check the format indicator
	if formatIndicator != formatIndicatorTCPWithoutSOAP {
		return fmt.Errorf("invalid format indicator")
	}

	// Read the message length
	var length uint16
	err = binary.Read(msgReader, binary.BigEndian, &length)
	if err != nil {
		return fmt.Errorf("failed to read message length: %w", err)
	}

	// Check the message length
	if uint16(len(data)) != length {
		return fmt.Errorf("invalid message length")
	}

	// Skip the invoke ID
	msgReader.Discard(4)

	// read the message body
	body, err := io.ReadAll(msgReader)
	if err != nil {
		return fmt.Errorf("failed to read message body: %w", err)
	}

	err = xml.Unmarshal(body, m)
	if err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	return nil
}
