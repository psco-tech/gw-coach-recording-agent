package csta

import (
	"encoding/xml"
	"fmt"
	"reflect"
)

const (
	MessageTypeStartApplicationSession                 MessageType = "StartApplicationSession"
	MessageTypeStopApplicationSession                  MessageType = "StopApplicationSession"
	MessageTypeStartApplicationSessionPosResponse      MessageType = "StartApplicationSessionPosResponse"
	MessageTypeStopApplicationSessionPosResponse       MessageType = "StopApplicationSessionPosResponse"
	MessageTypeStartApplicationSessionNegResponse      MessageType = "StartApplicationSessionNegResponse"
	MessageTypeResetApplicationSessionTimer            MessageType = "ResetApplicationSessionTimer"
	MessageTypeResetApplicationSessionTimerPosResponse MessageType = "ResetApplicationSessionTimerPosResponse"
	MessageTypeResetApplicationSessionTimerNegResponse MessageType = "ResetApplicationSessionTimerNegResponse"
)

func init() {
	registerMessageType(MessageTypeStartApplicationSession, reflect.TypeOf(StartApplicationSession{}))
	registerMessageType(MessageTypeStopApplicationSession, reflect.TypeOf(StopApplicationSession{}))
	registerMessageType(MessageTypeStartApplicationSessionPosResponse, reflect.TypeOf(StartApplicationSessionPosResponse{}))
	registerMessageType(MessageTypeStopApplicationSessionPosResponse, reflect.TypeOf(StopApplicationSessionPosResponse{}))
	registerMessageType(MessageTypeStartApplicationSessionNegResponse, reflect.TypeOf(StartApplicationSessionNegResponse{}))
	registerMessageType(MessageTypeResetApplicationSessionTimer, reflect.TypeOf(ResetApplicationSessionTimer{}))
	registerMessageType(MessageTypeResetApplicationSessionTimerPosResponse, reflect.TypeOf(ResetApplicationSessionTimerPosResponse{}))
	registerMessageType(MessageTypeResetApplicationSessionTimerNegResponse, reflect.TypeOf(ResetApplicationSessionTimerNegResponse{}))
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

type StopApplicationSession struct {
	XMLName          xml.Name `xml:"http://www.ecma-international.org/standards/ecma-354/appl_session StopApplicationSession"`
	SessionID        string   `xml:"sessionID"`
	SessionEndReason string   `xml:"sessionEndReason>appEndReason"`
}

func (m StopApplicationSession) Type() MessageType {
	return MessageTypeStopApplicationSession
}

type StopApplicationSessionPosResponse struct {
	XMLName xml.Name `xml:"http://www.ecma-international.org/standards/ecma-354/appl_session StopApplicationSessionPosResponse"`
}

func (StopApplicationSessionPosResponse) Type() MessageType {
	return MessageTypeStopApplicationSessionPosResponse
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

type ResetApplicationSessionTimer struct {
	XMLName                  xml.Name `xml:"http://www.ecma-international.org/standards/ecma-354/appl_session ResetApplicationSessionTimer"`
	SessionID                string   `xml:"sessionID"`
	RequestedSessionDuration uint     `xml:"requestedSessionDuration"`
}

func (ResetApplicationSessionTimer) Type() MessageType {
	return MessageTypeResetApplicationSessionTimer
}

type ResetApplicationSessionTimerPosResponse struct {
	XMLName               xml.Name `xml:"http://www.ecma-international.org/standards/ecma-354/appl_session ResetApplicationSessionTimerPosResponse"`
	ActualSessionDuration uint     `xml:"actualSessionDuration"`
}

func (ResetApplicationSessionTimerPosResponse) Type() MessageType {
	return MessageTypeResetApplicationSessionTimerPosResponse
}

type ResetApplicationSessionTimerNegResponse struct {
	XMLName xml.Name `xml:"http://www.ecma-international.org/standards/ecma-354/appl_session ResetApplicationSessionTimerNegResponse"`
}

func (ResetApplicationSessionTimerNegResponse) Type() MessageType {
	return MessageTypeResetApplicationSessionTimerNegResponse
}

func (c *cstaConn) StartApplicationSession(applicationId string, applicationSpecificInfo interface{}, protocolVersion string, callback ...HandleFunc) error {
	if c.state != ConnectionStateIdle {
		return fmt.Errorf("connection is not idle")
	}

	c.state = ConnectionStateStartingSession

	// Send out a StartApplicationSession request
	err := c.Request(StartApplicationSession{
		ApplicationID:            applicationId,
		RequestedSessionDuration: defaultSessionDuration,
		ProtocolVersion:          protocolVersion,
		ApplicationSpecificInfo:  applicationSpecificInfo,
	}, func(ctx *Context) {
		if ctx.Error == nil {
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
		}

		dispatchCallbacks(ctx, callback...)
	})

	if err != nil {
		return err
	}

	return nil
}
