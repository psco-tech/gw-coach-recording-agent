package csta

import (
	"encoding/xml"
	"reflect"
)

const (
	MessageTypeCSTAErrorCode MessageType = "CSTAErrorCode"
)

func init() {
	registerMessageType(MessageTypeCSTAErrorCode, reflect.TypeOf(CSTAErrorCode{}))
}

type CSTAErrorCode struct {
	XMLName   xml.Name `xml:"CSTAErrorCode"`
	Operation string   `xml:"operation,omitempty"`
}

func (CSTAErrorCode) Type() MessageType {
	return MessageTypeCSTAErrorCode
}
