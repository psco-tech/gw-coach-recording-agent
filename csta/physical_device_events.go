package csta

import (
	"encoding/xml"
	"reflect"
)

const (
	MessageTypeButtonInformationEvent MessageType = "ButtonInformationEvent"
)

func init() {
	registerMessageType(MessageTypeButtonInformationEvent, reflect.TypeOf(ButtonInformationEvent{}))
}

type ButtonInformationEvent struct {
	XMLName           xml.Name        `xml:"ButtonInformationEvent"`
	MonitorCrossRefID string          `xml:"monitorCrossRefID"`
	Device            SubjectDeviceID `xml:"device"`
	Button            ButtonID        `xml:"button"`
}

func (ButtonInformationEvent) Type() MessageType {
	return MessageTypeButtonInformationEvent
}
