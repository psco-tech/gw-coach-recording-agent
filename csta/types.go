package csta

type CSTAObject struct {
	DeviceObject *DeviceID     `xml:"deviceObject,omitempty"`
	CallObject   *ConnectionID `xml:"callObject,omitempty"`
}

type DeviceID struct {
	Device       string `xml:",chardata"`
	TypeOfNumber string `xml:"typeOfNumber,attr,omitempty"`
	// TODO implement the rest as needed
}

type ExtendedDeviceID struct {
	DeviceIdentifier DeviceID `xml:"deviceIdentifier"`
	NotKnown         *Empty   `xml:"notKnown,omitempty"`
	Restricted       *Empty   `xml:"restricted,omitempty"`
}

type ButtonID string

type Empty struct {
}

type SubjectDeviceID struct {
	ExtendedDeviceID
}

type CallingDeviceID struct {
	ExtendedDeviceID
}

type CalledDeviceID struct {
	ExtendedDeviceID
}

type LocalDeviceID struct {
	Device       string `xml:",chardata"`
	TypeOfNumber string `xml:"typeOfNumber,attr"`
	// TODO implement the rest as needed
}

type RedirectionDeviceID struct {
	NumberDialed *DeviceID `xml:"numberDialed,omitempty"`
	NotKnown     *Empty    `xml:"notKnown,omitempty"`
	Restricted   *Empty    `xml:"restricted,omitempty"`
	NotRequired  *Empty    `xml:"notRequired,omitempty"`
	NotSpecified *Empty    `xml:"notSpecified,omitempty"`
}

type ConnectionID struct {
	CallID   string         `xml:"callID,omitempty"`
	DeviceID *LocalDeviceID `xml:"deviceID,omitempty"`
}

type MonitorType string

const (
	MonitorTypeCall   MonitorType = "call"
	MonitorTypeDevice MonitorType = "device"
)

type Device struct {
	DeviceID *DeviceID `xml:"deviceID"`
	Category string    `xml:"deviceCategory"`
}

type DeviceList struct {
	Devices []Device `xml:"device"`
}
