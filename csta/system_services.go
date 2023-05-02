package csta

import (
	"encoding/xml"
	"reflect"
)

const (
	MessageTypeSystemStatus         MessageType = "SystemStatus"
	MessageTypeSystemStatusResponse MessageType = "SystemStatusResponse"
)

func init() {
	registerMessageType(MessageTypeSystemStatus, reflect.TypeOf(SystemStatus{}))
	registerMessageType(MessageTypeSystemStatusResponse, reflect.TypeOf(SystemStatusResponse{}))
}

type SystemStatus struct {
	XMLName xml.Name `xml:"SystemStatus"`
}

func (m SystemStatus) Type() MessageType {
	return MessageTypeSystemStatus
}

type SystemStatusResponse struct {
	XMLName xml.Name `xml:"SystemStatusResponse"`
}

func (m SystemStatusResponse) Type() MessageType {
	return MessageTypeSystemStatusResponse
}

func acknowledgeSystemStatus(c *Context) {
	c.conn.Write(c.InvokeID, SystemStatusResponse{})
}

func ignoreMessage(c *Context) {
}
