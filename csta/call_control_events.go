package csta

import (
	"encoding/xml"
	"reflect"
)

const (
	MessageTypeOriginatedEvent  MessageType = "OriginatedEvent"
	MessageTypeDeliveredEvent   MessageType = "DeliveredEvent"
	MessageTypeEstablishedEvent MessageType = "EstablishedEvent"
)

func init() {
	registerMessageType(MessageTypeOriginatedEvent, reflect.TypeOf(OriginatedEvent{}))
	registerMessageType(MessageTypeDeliveredEvent, reflect.TypeOf(DeliveredEvent{}))
	registerMessageType(MessageTypeEstablishedEvent, reflect.TypeOf(EstablishedEvent{}))
}

type OriginatedEvent struct {
	XMLName              xml.Name        `xml:"http://www.ecma-international.org/standards/ecma-323/csta/ed4 OriginatedEvent"`
	MonitorCrossRefID    string          `xml:"monitorCrossRefID"`
	OriginatedConnection ConnectionID    `xml:"originatedConnection"`
	CallingDevice        SubjectDeviceID `xml:"callingDevice"`
	CalledDevice         SubjectDeviceID `xml:"calledDevice"`
	LocalConnectionInfo  string          `xml:"localConnectionInfo"`
	Cause                string          `xml:"cause"`
}

func (OriginatedEvent) Type() MessageType {
	return MessageTypeOriginatedEvent
}

type DeliveredEvent struct {
	XMLName               xml.Name            `xml:"http://www.ecma-international.org/standards/ecma-323/csta/ed4 DeliveredEvent"`
	MonitorCrossRefID     string              `xml:"monitorCrossRefID"`
	Connection            ConnectionID        `xml:"connection"`
	AlertingDevice        SubjectDeviceID     `xml:"alertingDevice"`
	CallingDevice         SubjectDeviceID     `xml:"callingDevice"`
	CalledDevice          SubjectDeviceID     `xml:"calledDevice"`
	LastRedirectionDevice RedirectionDeviceID `xml:"lastRedirectionDevice"`
	LocalConnectionInfo   string              `xml:"localConnectionInfo"`
	Cause                 string              `xml:"cause"`
}

func (DeliveredEvent) Type() MessageType {
	return MessageTypeDeliveredEvent
}

type EstablishedEvent struct {
	XMLName               xml.Name            `xml:"http://www.ecma-international.org/standards/ecma-323/csta/ed4 EstablishedEvent"`
	MonitorCrossRefID     string              `xml:"monitorCrossRefID"`
	EstablishedConnection ConnectionID        `xml:"establishedConnection"`
	AnsweringDevice       SubjectDeviceID     `xml:"answeringDevice"`
	CallingDevice         SubjectDeviceID     `xml:"callingDevice"`
	CalledDevice          SubjectDeviceID     `xml:"calledDevice"`
	LastRedirectionDevice RedirectionDeviceID `xml:"lastRedirectionDevice"`
	LocalConnectionInfo   string              `xml:"localConnectionInfo"`
	Cause                 string              `xml:"cause"`
}

func (EstablishedEvent) Type() MessageType {
	return MessageTypeEstablishedEvent
}
