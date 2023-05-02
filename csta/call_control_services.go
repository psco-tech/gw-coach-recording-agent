package csta

import "reflect"

const (
	MessageTypeMakeCall MessageType = "MakeCall"
)

func init() {
	registerMessageType(MessageTypeMakeCall, reflect.TypeOf(MakeCall{}))
}

type MakeCall struct {
	CallingDevice         string `xml:"callingDevice"`
	CalledDirectoryNumber string `xml:"calledDirectoryNumber"`
	AutoOriginate         string `xml:"autoOriginate,omitempty"`
}

func (MakeCall) Type() MessageType {
	return MessageTypeMakeCall
}
