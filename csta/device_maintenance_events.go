package csta

import (
	"encoding/xml"
	"reflect"
)

const (
	MessageTypeOutOfServiceEvent  MessageType = "OutOfServiceEvent"
	MessageTypeBackInServiceEvent MessageType = "BackInServiceEvent"
)

func init() {
	registerMessageType(MessageTypeOutOfServiceEvent, reflect.TypeOf(OutOfServiceEvent{}))
	registerMessageType(MessageTypeBackInServiceEvent, reflect.TypeOf(BackInServiceEvent{}))
}

type OutOfServiceEvent struct {
	XMLName           xml.Name        `xml:"OutOfServiceEvent"`
	MonitorCrossRefID string          `xml:"monitorCrossRefID"`
	Device            SubjectDeviceID `xml:"device"`
}

func (OutOfServiceEvent) Type() MessageType {
	return MessageTypeOutOfServiceEvent
}

type BackInServiceEvent struct {
	XMLName           xml.Name        `xml:"BackInServiceEvent"`
	MonitorCrossRefID string          `xml:"monitorCrossRefID"`
	Device            SubjectDeviceID `xml:"device"`
}

func (BackInServiceEvent) Type() MessageType {
	return MessageTypeBackInServiceEvent
}
