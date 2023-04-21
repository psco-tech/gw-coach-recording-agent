package csta

import (
	"encoding/xml"
	"reflect"
)

const (
	MessageTypeServiceInitiatedEvent  MessageType = "ServiceInitiatedEvent"
	MessageTypeOriginatedEvent        MessageType = "OriginatedEvent"
	MessageTypeDeliveredEvent         MessageType = "DeliveredEvent"
	MessageTypeEstablishedEvent       MessageType = "EstablishedEvent"
	MessageTypeConnectionClearedEvent MessageType = "ConnectionClearedEvent"
)

func init() {
	registerMessageType(MessageTypeServiceInitiatedEvent, reflect.TypeOf(ServiceInitiatedEvent{}))
	registerMessageType(MessageTypeOriginatedEvent, reflect.TypeOf(OriginatedEvent{}))
	registerMessageType(MessageTypeDeliveredEvent, reflect.TypeOf(DeliveredEvent{}))
	registerMessageType(MessageTypeEstablishedEvent, reflect.TypeOf(EstablishedEvent{}))
	registerMessageType(MessageTypeConnectionClearedEvent, reflect.TypeOf(ConnectionClearedEvent{}))
}

type ServiceInitiatedEvent struct {
	XMLName             xml.Name         `xml:"ServiceInitiatedEvent"`
	MonitorCrossRefID   string           `xml:"monitorCrossRefID"`
	InitiatedConnection ConnectionID     `xml:"initiatedConnection"`
	InititatingDevice   SubjectDeviceID  `xml:"initiatingDevice"`
	LocalConnectionInfo string           `xml:"localConnectionInfo"`
	Cause               string           `xml:"cause"`
	CallLinkageData     *CallLinkageData `xml:"callLinkageData,omitempty"`
}

func (ServiceInitiatedEvent) Type() MessageType {
	return MessageTypeServiceInitiatedEvent
}

type OriginatedEvent struct {
	XMLName              xml.Name        `xml:"OriginatedEvent"`
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
	XMLName               xml.Name            `xml:"DeliveredEvent"`
	MonitorCrossRefID     string              `xml:"monitorCrossRefID"`
	Connection            ConnectionID        `xml:"connection"`
	AlertingDevice        SubjectDeviceID     `xml:"alertingDevice"`
	CallingDevice         SubjectDeviceID     `xml:"callingDevice"`
	CalledDevice          SubjectDeviceID     `xml:"calledDevice"`
	LastRedirectionDevice RedirectionDeviceID `xml:"lastRedirectionDevice"`
	LocalConnectionInfo   string              `xml:"localConnectionInfo"`
	CallLinkageData       *CallLinkageData    `xml:"callLinkageData,omitempty"`
	Cause                 string              `xml:"cause"`
}

func (DeliveredEvent) Type() MessageType {
	return MessageTypeDeliveredEvent
}

type EstablishedEvent struct {
	XMLName               xml.Name            `xml:"EstablishedEvent"`
	MonitorCrossRefID     string              `xml:"monitorCrossRefID"`
	EstablishedConnection ConnectionID        `xml:"establishedConnection"`
	AnsweringDevice       SubjectDeviceID     `xml:"answeringDevice"`
	CallingDevice         SubjectDeviceID     `xml:"callingDevice"`
	CalledDevice          SubjectDeviceID     `xml:"calledDevice"`
	LastRedirectionDevice RedirectionDeviceID `xml:"lastRedirectionDevice"`
	LocalConnectionInfo   string              `xml:"localConnectionInfo"`
	Cause                 string              `xml:"cause"`
	CallLinkageData       *CallLinkageData    `xml:"callLinkageData,omitempty"`
}

func (EstablishedEvent) Type() MessageType {
	return MessageTypeEstablishedEvent
}

type ConnectionClearedEvent struct {
	XMLName             xml.Name        `xml:"ConnectionClearedEvent"`
	MonitorCrossRefID   string          `xml:"monitorCrossRefID"`
	DroppedConnection   ConnectionID    `xml:"droppedConnection"`
	ReleasingDevice     SubjectDeviceID `xml:"releasingDevice"`
	LocalConnectionInfo string          `xml:"localConnectionInfo"`
	Cause               string          `xml:"cause"`
}

func (ConnectionClearedEvent) Type() MessageType {
	return MessageTypeConnectionClearedEvent
}
