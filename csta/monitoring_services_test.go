package csta

import (
	"bytes"
	"encoding/hex"
	"testing"
)

var monitorStartMessage = []byte("\x00\x00\x00\xEB0001<MonitorStart xmlns=\"http://www.ecma-international.org/standards/ecma-323/csta/ed4\"><monitorObject><deviceObject typeOfNumber=\"dialingNumber\">212700</deviceObject></monitorObject><monitorType>device</monitorType></MonitorStart>")

func TestMarshalMonitorStart(t *testing.T) {
	m := &MonitorStart{
		MonitorObject: CSTAObject{DeviceObject: &DeviceID{
			Device:       "212700",
			TypeOfNumber: "dialingNumber",
		}},
		MonitorType: MonitorTypeDevice,
	}

	marshalledMessage, err := marshal(1, m)
	if err != nil {
		t.Error(err)
	}

	if !bytes.Equal(marshalledMessage, monitorStartMessage) {
		t.Logf("\n%s\n", hex.Dump(marshalledMessage))
		t.Fail()
	}
}

func TestUnmarshalMonitorStart(t *testing.T) {
	msg := MonitorStart{}
	err := unmarshal(monitorStartMessage, &msg)

	if err != nil {
		t.Log(err)
		t.Fail()
	}

	if msg.MonitorObject.DeviceObject.TypeOfNumber != "dialingNumber" || msg.MonitorObject.DeviceObject.Device != "212700" {
		t.Fail()
	}
}
