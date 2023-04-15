package csta

import (
	"encoding/xml"
	"fmt"
	"reflect"
)

const (
	MessageTypeStartApplicationSession            MessageType = "StartApplicationSession"
	MessageTypeStartApplicationSessionPosResponse MessageType = "StartApplicationSessionPosResponse"
	MessageTypeStartApplicationSessionNegResponse MessageType = "StartApplicationSessionNegResponse"
)

func init() {
	registerMessageType(MessageTypeStartApplicationSession, reflect.TypeOf(StartApplicationSession{}))
	registerMessageType(MessageTypeStartApplicationSessionPosResponse, reflect.TypeOf(StartApplicationSessionPosResponse{}))
	registerMessageType(MessageTypeStartApplicationSessionNegResponse, reflect.TypeOf(StartApplicationSessionNegResponse{}))
}

type StartApplicationSession struct {
	XMLName                  xml.Name    `xml:"http://www.ecma-international.org/standards/ecma-354/appl_session StartApplicationSession"`
	ApplicationID            string      `xml:"applicationInfo>applicationID"`
	ApplicationSpecificInfo  interface{} `xml:"applicationInfo>applicationSpecificInfo,omitempty"`
	ProtocolVersion          string      `xml:"requestedProtocolVersions>protocolVersion"`
	RequestedSessionDuration uint        `xml:"requestedSessionDuration"`
}

func (m StartApplicationSession) Type() MessageType {
	return MessageTypeStartApplicationSession
}

type StartApplicationSessionPosResponse struct {
	XMLName               xml.Name `xml:"http://www.ecma-international.org/standards/ecma-354/appl_session StartApplicationSessionPosResponse"`
	SessionID             string   `xml:"sessionID"`
	ActualProtocolVersion string   `xml:"actualProtocolVersion"`
	ActualSessionDuration uint     `xml:"actualSessionDuration"`
}

func (m StartApplicationSessionPosResponse) Type() MessageType {
	return MessageTypeStartApplicationSessionPosResponse
}

type StartApplicationSessionNegResponse struct {
	XMLName xml.Name `xml:"http://www.ecma-international.org/standards/ecma-354/appl_session StartApplicationSessionNegResponse"`
}

func (m StartApplicationSessionNegResponse) Type() MessageType {
	return MessageTypeStartApplicationSessionNegResponse
}

func (c *cstaConn) StartApplicationSession(applicationId string, applicationSpecificInfo interface{}, callback ...HandleFunc) error {
	if c.state != ConnectionStateIdle {
		return fmt.Errorf("connection is not idle")
	}

	c.state = ConnectionStateStartingSession

	// Send out a StartApplicationSession request
	err := c.Request(StartApplicationSession{
		ApplicationID:            applicationId,
		RequestedSessionDuration: defaultSessionDuration,
		ProtocolVersion:          "http://www.ecma-international.org/standards/ecma-323/csta/ed4",
		ApplicationSpecificInfo:  applicationSpecificInfo,
	}, func(ctx *Context) {
		switch ctx.Message.Type() {
		case MessageTypeStartApplicationSessionPosResponse:
			c.sessionId = ctx.Message.(*StartApplicationSessionPosResponse).SessionID
			c.state = ConnectionStateActive
			// TODO start periodic refresh

		case MessageTypeStartApplicationSessionNegResponse:
			c.state = ConnectionStateError
			c.Close()
			ctx.Error = fmt.Errorf("received StartApplicationSessionNegResponse")
		}

		dispatchCallbacks(ctx, callback...)
	})

	if err != nil {
		return err
	}

	return nil
}
