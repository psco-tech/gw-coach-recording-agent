package csta

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"testing"
)

func TestRead(t *testing.T) {
	mockData := bytes.NewBuffer(startApplicationSessionMessage)
	conn := cstaConn{
		rw: bufio.NewReadWriter(bufio.NewReader(mockData), bufio.NewWriter(mockData)),
	}

	invokeId, message, err := conn.Read()
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if invokeId != 1 {
		t.Fail()
	}
	if _, ok := message.(*StartApplicationSession); !ok {
		t.Fail()
	}
}

func TestWrite(t *testing.T) {
	mockData := new(bytes.Buffer)
	conn := cstaConn{
		options: &ConnectionOptions{},
		rw:      bufio.NewReadWriter(bufio.NewReader(mockData), bufio.NewWriter(mockData)),
	}
	err := conn.Write(1, StartApplicationSession{
		ApplicationID:            "testApplicationId",
		ProtocolVersion:          "http://www.ecma-international.org/standards/ecma-323/csta/ed4",
		RequestedSessionDuration: 300,
	})
	if err != nil {
		t.Log(err)
		t.Fail()
	}
	if !bytes.Equal(mockData.Bytes(), startApplicationSessionMessage) {
		t.Logf("%s\n", hex.Dump(mockData.Bytes()))
		t.Fail()
	}
}

func TestRequest(t *testing.T) {
	mockData := new(bytes.Buffer)
	conn := cstaConn{
		options:      &ConnectionOptions{},
		rw:           bufio.NewReadWriter(bufio.NewReader(mockData), bufio.NewWriter(mockData)),
		transactions: make(map[uint]HandleFunc),
	}

	conn.Request(StartApplicationSession{
		ApplicationID:            "testApplicationId",
		ProtocolVersion:          "http://www.ecma-international.org/standards/ecma-323/csta/ed4",
		RequestedSessionDuration: 300,
	}, func(c *Context) {
		// Ignore
	})

	if len(conn.transactions) != 1 {
		t.Fail()
	}
}

func TestStartApplicationSession(t *testing.T) {
	mockData := new(bytes.Buffer)
	conn := cstaConn{
		options:      &ConnectionOptions{},
		rw:           bufio.NewReadWriter(bufio.NewReader(mockData), bufio.NewWriter(mockData)),
		transactions: make(map[uint]HandleFunc),
	}
	err := conn.StartApplicationSession("testApplicationId", struct {
		Username string `xml:"userName"`
		Password string `xml:"password"`
	}{
		Username: "username",
		Password: "password",
	}, "http://www.ecma-international.org/standards/ecma-323/csta/ed4")

	if err != nil {
		t.Log(err)
		t.Fail()
	}
}

func TestClose(t *testing.T) {
	mockData := new(bytes.Buffer)
	conn := cstaConn{
		options:      &ConnectionOptions{},
		rw:           bufio.NewReadWriter(bufio.NewReader(mockData), bufio.NewWriter(mockData)),
		transactions: make(map[uint]HandleFunc),
	}
	conn.Close()
}
