package csta

import (
	"encoding/xml"
	"reflect"
)

const (
	MessageTypeGetPhysicalDeviceInformation         MessageType = "GetPhysicalDeviceInformation"
	MessageTypeGetPhysicalDeviceInformationResponse MessageType = "GetPhysicalDeviceInformationResponse"
	MessageTypeGetSwitchingFunctionDevices          MessageType = "GetSwitchingFunctionDevices"
	MessageTypeGetSwitchingFunctionDevicesResponse  MessageType = "GetSwitchingFunctionDevicesResponse"
	MessageTypeSwitchingFunctionDevices             MessageType = "SwitchingFunctionDevices"
)

func init() {
	registerMessageType(MessageTypeGetPhysicalDeviceInformation, reflect.TypeOf(GetPhysicalDeviceInformation{}))
	registerMessageType(MessageTypeGetPhysicalDeviceInformationResponse, reflect.TypeOf(GetPhysicalDeviceInformationResponse{}))
	registerMessageType(MessageTypeGetSwitchingFunctionDevices, reflect.TypeOf(GetSwitchingFunctionDevices{}))
	registerMessageType(MessageTypeGetSwitchingFunctionDevicesResponse, reflect.TypeOf(GetSwitchingFunctionDevicesResponse{}))
	registerMessageType(MessageTypeSwitchingFunctionDevices, reflect.TypeOf(SwitchingFunctionDevices{}))

}

type GetPhysicalDeviceInformation struct {
	XMLName xml.Name `xml:"GetPhysicalDeviceInformation"`
	Device  DeviceID `xml:"device"`
}

func (m GetPhysicalDeviceInformation) Type() MessageType {
	return MessageTypeGetPhysicalDeviceInformation
}

type GetPhysicalDeviceInformationResponse struct {
	XMLName xml.Name `xml:"GetPhysicalDeviceInformationResponse"`
}

func (m GetPhysicalDeviceInformationResponse) Type() MessageType {
	return MessageTypeGetPhysicalDeviceInformationResponse
}

type GetSwitchingFunctionDevices struct {
	XMLName xml.Name `xml:"GetSwitchingFunctionDevices"`
}

func (m GetSwitchingFunctionDevices) Type() MessageType {
	return MessageTypeGetSwitchingFunctionDevices
}

type GetSwitchingFunctionDevicesResponse struct {
	XMLName xml.Name `xml:"GetSwitchingFunctionDevicesResponse"`
}

func (m GetSwitchingFunctionDevicesResponse) Type() MessageType {
	return MessageTypeGetSwitchingFunctionDevicesResponse
}

type SwitchingFunctionDevices struct {
	XMLName           xml.Name   `xml:"SwitchingFunctionDevices"`
	ServiceCrossRefID string     `xml:"serviceCrossRefID"`
	SegmentID         uint       `xml:"segmentID"`
	LastSegment       bool       `xml:"lastSegment"`
	DeviceList        DeviceList `xml:"deviceList"`
}

func (m SwitchingFunctionDevices) Type() MessageType {
	return MessageTypeSwitchingFunctionDevices
}
