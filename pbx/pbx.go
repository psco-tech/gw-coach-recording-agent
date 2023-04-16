package pbx

import (
	"context"
	"errors"

	"github.com/psco-tech/gw-coach-recording-agent/csta"
	"github.com/spf13/viper"
)

var implementations map[string]PBX

func init() {
	viper.SetDefault("pbx_type", "osbiz")
	viper.SetDefault("pbx_address", "192.168.1.30:8800")
	viper.SetDefault("pbx_username", "AMHOST")
	viper.SetDefault("pbx_password", "77777")
}

type ConnectionState int

const (
	ConnectionStateError        ConnectionState = -1
	ConnectionStateDisconnected ConnectionState = 0
	ConnectionStateConnected    ConnectionState = 1
)

type PBX interface {
	Connect() (csta.Conn, error)

	ConnectionState() ConnectionState
	MonitorStart(deviceId string) (monitorPoint MonitorPoint, err error)
	SetContext(ctx context.Context)
	Serve() error
	Close() error
}

func RegisterImplementation(id string, implementation PBX) {
	if implementations == nil {
		implementations = make(map[string]PBX)
	}

	implementations[id] = implementation
}

func New(implementationId string, ctx context.Context) (PBX, error) {
	if implementations == nil {
		return nil, errors.New("no PBX implementations registered")
	}

	if impl, ok := implementations[implementationId]; ok {
		impl.SetContext(ctx)
		return impl, nil
	}
	return nil, errors.New("unknown PBX implementation ID")
}

type Connection struct {
}
