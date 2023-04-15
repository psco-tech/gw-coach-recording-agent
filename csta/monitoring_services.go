package csta

import (
	"encoding/xml"
	"reflect"
)

const (
	MessageTypeMonitorStart         MessageType = "MonitorStart"
	MessageTypeMonitorStartResponse MessageType = "MonitorStartResponse"
)

func init() {
	registerMessageType(MessageTypeMonitorStart, reflect.TypeOf(MonitorStart{}))
	registerMessageType(MessageTypeMonitorStartResponse, reflect.TypeOf(MonitorStartResponse{}))
}

type MonitorStart struct {
	XMLName       xml.Name    `xml:"http://www.ecma-international.org/standards/ecma-323/csta/ed4 MonitorStart"`
	MonitorObject CSTAObject  `xml:"monitorObject"`
	MonitorType   MonitorType `xml:"monitorType"`
}

func (m MonitorStart) Type() MessageType {
	return MessageTypeMonitorStart
}

type MonitorStartResponse struct {
	XMLName           xml.Name `xml:"http://www.ecma-international.org/standards/ecma-323/csta/ed4 MonitorStartResponse"`
	MonitorCrossRefID string   `xml:"monitorCrossRefID"`
}

func (m MonitorStartResponse) Type() MessageType {
	return MessageTypeMonitorStartResponse
}

func (c *cstaConn) MonitorStart(monitorObject CSTAObject, monitorType MonitorType, callback ...HandleFunc) error {
	return c.Request(MonitorStart{
		MonitorObject: monitorObject,
		MonitorType:   monitorType,
	}, func(c *Context) {
		dispatchCallbacks(c, callback...)
	})
}
