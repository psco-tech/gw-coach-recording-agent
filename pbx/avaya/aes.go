package avaya

import (
	"github.com/psco-tech/gw-coach-recording-agent/csta"
	"github.com/psco-tech/gw-coach-recording-agent/pbx"
)

type AvayaAES struct {
	conn csta.Conn
}

func (aes *AvayaAES) Connect(network, addr, applicationId, username, password string) error {
	// TODO implement
	return nil
}

func (aes *AvayaAES) ConnectionState() pbx.ConnectionState {
	switch aes.conn.State() {
	case csta.ConnectionStateActive:
		return pbx.ConnectionStateConnected
	case csta.ConnectionStateError:
		return pbx.ConnectionStateError
	}

	return pbx.ConnectionStateDisconnected
}

func (aes *AvayaAES) Close() error {
	return aes.conn.Close()
}

func (aes *AvayaAES) MonitorStart(deviceId string) (monitorPoint pbx.MonitorPoint, err error) {
	// TODO implement
	return nil, nil
}
