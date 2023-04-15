package pbx

import "github.com/psco-tech/gw-coach-recording-agent/csta"

type MonitorPoint interface {
	Device() Device
	CrossReferenceID() string
	Events() <-chan csta.Message
}
