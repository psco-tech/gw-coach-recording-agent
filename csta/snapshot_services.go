package csta

import (
	"encoding/xml"
	"reflect"
)

const (
	MessageTypeSnapshotDevice         MessageType = "SnapshotDevice"
	MessageTypeSnapshotDeviceResponse MessageType = "SnapshotDeviceResponse"
)

func init() {
	registerMessageType(MessageTypeSnapshotDevice, reflect.TypeOf(SnapshotDevice{}))
	registerMessageType(MessageTypeSnapshotDeviceResponse, reflect.TypeOf(SnapshotDeviceResponse{}))
}

type SnapshotDevice struct {
	XMLName        xml.Name   `xml:"SnapshotDevice"`
	SnapshotObject CSTAObject `xml:"snapshotObject"`
}

func (SnapshotDevice) Type() MessageType {
	return MessageTypeSnapshotDevice
}

type SnapshotDeviceResponse struct {
	XMLName           xml.Name `xml:"SnapshotDeviceResponse"`
	ServiceCrossRefID string   `xml:"serviceCrossRefID"`
}

func (SnapshotDeviceResponse) Type() MessageType {
	return MessageTypeSnapshotDeviceResponse
}
