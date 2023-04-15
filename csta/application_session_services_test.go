package csta

import (
	"bytes"
	"encoding/hex"
	"encoding/xml"
	"testing"
)

var startApplicationSessionMessage = []byte("\x00\x00\x01\xA70001<StartApplicationSession xmlns=\"http://www.ecma-international.org/standards/ecma-354/appl_session\"><applicationInfo><applicationID>testApplicationId</applicationID></applicationInfo><requestedProtocolVersions><protocolVersion>http://www.ecma-international.org/standards/ecma-323/csta/ed4</protocolVersion></requestedProtocolVersions><requestedSessionDuration>300</requestedSessionDuration></StartApplicationSession>")
var startApplicationSessionMessageWithApplicationSpecificInfo = []byte("\x00\x00\x02\x9f0002<StartApplicationSession xmlns=\"http://www.ecma-international.org/standards/ecma-354/appl_session\"><applicationInfo><applicationID>testApplicationId</applicationID><applicationSpecificInfo><SessionLoginInfo xmlns=\"http://www.avaya.com/csta\"><userName>username</userName><password>password</password><sessionCleanupDelay>60</sessionCleanupDelay><sessionID></sessionID></SessionLoginInfo></applicationSpecificInfo></applicationInfo><requestedProtocolVersions><protocolVersion>http://www.ecma-international.org/standards/ecma-323/csta/ed4</protocolVersion></requestedProtocolVersions><requestedSessionDuration>300</requestedSessionDuration></StartApplicationSession>")

func TestMarshalStartApplicationSession(t *testing.T) {
	msg := StartApplicationSession{
		ApplicationID:            "testApplicationId",
		ProtocolVersion:          "http://www.ecma-international.org/standards/ecma-323/csta/ed4",
		RequestedSessionDuration: 300,
	}
	marshalledMessage, err := marshal(1, msg)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(marshalledMessage, startApplicationSessionMessage) {
		t.Logf("\n%s\n", hex.Dump(marshalledMessage))
		t.Fail()
	}
}

func TestMarshalStartApplicationSessionWithApplicationSpecificInfo(t *testing.T) {
	applicationSpecificInfo := struct {
		SessionLoginInfo interface{}
	}{
		SessionLoginInfo: struct {
			XMLName             xml.Name `xml:"http://www.avaya.com/csta SessionLoginInfo"`
			Username            string   `xml:"userName"`
			Password            string   `xml:"password"`
			SessionCleanupDelay uint     `xml:"sessionCleanupDelay"`
			SessionID           string   `xml:"sessionID"`
		}{
			Username:            "username",
			Password:            "password",
			SessionCleanupDelay: 60,
			SessionID:           "",
		},
	}
	msg := StartApplicationSession{
		ApplicationID:            "testApplicationId",
		ProtocolVersion:          "http://www.ecma-international.org/standards/ecma-323/csta/ed4",
		ApplicationSpecificInfo:  applicationSpecificInfo,
		RequestedSessionDuration: 300,
	}
	marshalledMessage, err := marshal(2, msg)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(marshalledMessage, startApplicationSessionMessageWithApplicationSpecificInfo) {
		t.Logf("\n%s\n", hex.Dump(marshalledMessage))
		t.Fail()
	}
}

func TestUnmarshalStartApplicationSession(t *testing.T) {
	msg := StartApplicationSession{}
	err := unmarshal(startApplicationSessionMessage, &msg)

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	if msg.ApplicationID != "testApplicationId" || msg.ProtocolVersion != "http://www.ecma-international.org/standards/ecma-323/csta/ed4" {
		t.Fail()
	}
}
