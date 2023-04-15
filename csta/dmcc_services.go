package csta

import (
	"encoding/xml"
	"reflect"
)

const (
	MessageTypeRegisterTerminalRequest  MessageType = "RegisterTerminalRequest"
	MessageTypeRegisterTerminalResponse MessageType = "RegisterTerminalResponse"
	MessageTypeGetDeviceId              MessageType = "GetDeviceId"
	MessageTypeGetDeviceIdResponse      MessageType = "GetDeviceIdResponse"
)

func init() {
	registerMessageType(MessageTypeRegisterTerminalRequest, reflect.TypeOf(RegisterTerminalRequest{}))
	registerMessageType(MessageTypeRegisterTerminalResponse, reflect.TypeOf(RegisterTerminalResponse{}))
	registerMessageType(MessageTypeGetDeviceId, reflect.TypeOf(GetDeviceId{}))
	registerMessageType(MessageTypeGetDeviceIdResponse, reflect.TypeOf(GetDeviceIdResponse{}))
}

type LoginInfo struct {
	ForceLogin     bool            `xml:"forceLogin"`
	SharedControl  bool            `xml:"sharedControl"`
	Password       string          `xml:"password"`
	MediaMode      MediaMode       `xml:"mediaMode"`
	DependencyMode DependencyMode  `xml:"dependencyMode"`
	LocalMediaInfo *LocalMediaInfo `xml:"localMediaInfo,omitempty"`
}

type MediaMode string

const (
	MediaModeServer MediaMode = "SERVER"
	MediaModeClient MediaMode = "CLIENT"
)

type DependencyMode string

const (
	DependencyModeMain        DependencyMode = "MAIN"
	DependencyModeIndependent DependencyMode = "INDEPENDENT"
	DependencyModeDependent   DependencyMode = "DEPENDENT"
)

type LocalMediaInfo struct {
	RTPAddress     *NetworkEndpoint `xml:"rtpAddress,omitempty"`
	RTCPAddress    *NetworkEndpoint `xml:"rtcpAddress,omitempty"`
	Codecs         []string         `xml:"codecs"`
	PacketSize     int              `xml:"packetSize"`
	EncryptionList []string         `xml:"encryptionList"`
}

type NetworkEndpoint struct {
	Address string `xml:"address"`
	Port    int    `xml:"port"`
}

type RegisterTerminalRequest struct {
	XMLName        xml.Name        `xml:"http://www.avaya.com/csta RegisterTerminalRequest"`
	Device         DeviceID        `xml:"device"`
	LoginInfo      LoginInfo       `xml:"loginInfo"`
	LocalMediaInfo *LocalMediaInfo `xml:"localMediaInfo,omitempty"`
}

func (RegisterTerminalRequest) Type() MessageType {
	return MessageTypeRegisterTerminalRequest
}

type RegisterTerminalResponse struct {
	XMLName xml.Name `xml:"http://www.avaya.com/csta RegisterTerminalResponse"`
	Device  struct {
		Device DeviceID `xml:"deviceIdentifier"`
	} `xml:"device"`
	SignallingEncryption string `xml:"signallingEncryption"`
	Code                 string `xml:"code"`
}

func (RegisterTerminalResponse) Type() MessageType {
	return MessageTypeRegisterTerminalResponse
}

type GetDeviceId struct {
	XMLName           xml.Name `xml:"http://www.avaya.com/csta GetDeviceId"`
	SwitchName        string   `xml:"switchName,omitempty"`
	SwitchIPInterface string   `xml:"switchIÜInterface,omitempty"`
	Extension         string   `xml:"extension"`
}

func (GetDeviceId) Type() MessageType {
	return MessageTypeGetDeviceId
}

type GetDeviceIdResponse struct {
	XMLName xml.Name `xml:"http://www.avaya.com/csta GetDeviceIdResponse"`
	Device  DeviceID `xml:"device"`
}

func (GetDeviceIdResponse) Type() MessageType {
	return MessageTypeGetDeviceIdResponse
}
