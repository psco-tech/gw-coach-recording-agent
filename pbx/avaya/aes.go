package avaya

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"
	"time"

	"github.com/psco-tech/gw-coach-recording-agent/csta"
	"github.com/psco-tech/gw-coach-recording-agent/pbx"
	"github.com/spf13/viper"
)

const defaultEventBufferSize = 100

type AvayaAES struct {
	ctx           context.Context
	sessionId     string
	conn          csta.Conn
	monitorPoints map[string]*monitorPoint
}

func (aes *AvayaAES) SetContext(ctx context.Context) {
	aes.ctx = ctx
}

// Connect dials the connection and establishes an application session
func (aes *AvayaAES) Connect(network, addr, applicationId, username, password string) (csta.Conn, error) {
	cstaConn, err := csta.Dial(network, addr, aes.ctx, nil)
	if err != nil {
		return nil, err
	}

	aes.setupHandlers(cstaConn)

	var wg sync.WaitGroup

	aes.conn = cstaConn

	wg.Add(1)
	err = cstaConn.StartApplicationSession(applicationId, struct {
		SessionLoginInfo struct {
			Username            string `xml:"userName"`
			Password            string `xml:"password"`
			SessionCleanupDelay int    `xml:"sessionCleanupDelay"`
		} `xml:"SessionLoginInfo"`
	}{
		SessionLoginInfo: struct {
			Username            string "xml:\"userName\""
			Password            string "xml:\"password\""
			SessionCleanupDelay int    "xml:\"sessionCleanupDelay\""
		}{
			Username:            username,
			Password:            password,
			SessionCleanupDelay: 60,
		},
	},
		"http://www.ecma-international.org/standards/ecma-323/csta/ed3/priv5",
		func(ctx *csta.Context) {
			if ctx.Error != nil {
				err = ctx.Error
			}
			if r, ok := ctx.Message.(*csta.StartApplicationSessionPosResponse); ok {
				log.Printf("Application session started with session id <%s>\n", r.SessionID)
				aes.sessionId = r.SessionID
			}
			wg.Done()
		})

	if err != nil {
		return nil, err
	}

	wg.Wait()

	go aes.applicationSessionTimer(aes.ctx)

	return cstaConn, nil
}

// applicationSessionTimer handles resetting the application session timer every so often
// to prevent sessions from expiring
func (aes *AvayaAES) applicationSessionTimer(ctx context.Context) {
	timer := time.NewTicker(30 * time.Second)

	for {
		select {
		case <-timer.C:
			log.Printf("Resetting application session timer\n")
			aes.conn.Request(csta.ResetApplicationSessionTimer{
				SessionID:                aes.sessionId,
				RequestedSessionDuration: 60,
			}, func(c *csta.Context) {
				// TODO handle negative response
			})
		case <-aes.conn.Closed():
			timer.Stop()
			log.Printf("Conection closed, stopping application session timer\n")
			return
		case <-ctx.Done():
			timer.Stop()
			log.Printf("Application shutdown, stopping application session timer\n")
			return
		}
	}
}

func (aes *AvayaAES) setupHandlers(conn csta.Conn) {

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

// Close closes the TCP connection after it stopped the application session
func (aes *AvayaAES) Close() error {

	if aes.conn.State() == csta.ConnectionStateActive {
		var wg sync.WaitGroup
		wg.Add(1)

		aes.conn.Request(csta.StopApplicationSession{
			SessionID:        aes.sessionId,
			SessionEndReason: "Application Shutdown",
		}, func(c *csta.Context) {
			wg.Done()
		})
		wg.Wait()
	}

	return aes.conn.Close()
}

// MonitorStart gets hold of a device ID and calls MonitorStart on it
func (aes *AvayaAES) MonitorStart(extension string) (mp pbx.MonitorPoint, err error) {
	deviceId, err := aes.GetDeviceID(extension)
	if err != nil {
		return nil, err
	}

	var wg sync.WaitGroup
	wg.Add(1)

	aes.conn.MonitorStart(
		csta.CSTAObject{
			DeviceObject: &csta.DeviceID{
				Device:       deviceId,
				TypeOfNumber: "other",
				MediaClass:   "notKnown",
			},
		},
		csta.MonitorTypeDevice,
		func(c *csta.Context) {
			defer wg.Done()

			if c.Error != nil {
				err = c.Error
				return
			}

			if resp, ok := (c.Message).(*csta.MonitorStartResponse); ok {
				if aes.monitorPoints == nil {
					aes.monitorPoints = make(map[string]*monitorPoint)
				}
				aes.monitorPoints[resp.MonitorCrossRefID] = &monitorPoint{
					crossReferenceID: resp.MonitorCrossRefID,
					device: &device{
						extension: extension,
						deviceId:  deviceId,
					},
				}
				mp = aes.monitorPoints[resp.MonitorCrossRefID]
				return
			} else {
				err = fmt.Errorf("failed to start monitoring for extension <%s>", extension)
			}
		},
	)
	wg.Wait()

	if err == nil && mp != nil {
		log.Printf("Monitoring <%s (%s)> with CrossRefID <%s>\n", extension, deviceId, mp.CrossReferenceID())
	}

	return
}

// GetDeviceID gets the internal device ID for an extension
func (aes *AvayaAES) GetDeviceID(extension string) (deviceId string, err error) {
	var wg sync.WaitGroup
	wg.Add(1)

	aes.conn.Request(csta.GetDeviceId{
		SwitchName: viper.GetString("pbx_switch_name"),
		Extension:  extension,
	}, func(c *csta.Context) {
		defer wg.Done()
		if r, ok := c.Message.(*csta.GetDeviceIdResponse); ok {
			deviceId = r.Device.Device
		} else {
			err = fmt.Errorf("failed to get device id for extension <%s>", extension)
		}
	})
	wg.Wait()

	if len(deviceId) == 0 {
		err = fmt.Errorf("failed to get device id for extension <%s>", extension)
	}

	return
}

// RegisterTerminal will force-register a virtual station and instruct the Gateway to
// send any audio data to the specified local endpoint
func (aes *AvayaAES) RegisterTerminal(extension string, password string, localRtpEndpoint *net.UDPAddr) error {
	// Get the actual device ID for this extension
	deviceId, err := aes.GetDeviceID(extension)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup
	wg.Add(1)
	aes.conn.Request(csta.RegisterTerminalRequest{
		Device: csta.DeviceID{Device: deviceId, TypeOfNumber: "other", MediaClass: "notKnown"},
		LoginInfo: csta.LoginInfo{
			ForceLogin:     true,
			SharedControl:  false,
			Password:       password,
			MediaMode:      csta.MediaModeClient,
			DependencyMode: csta.DependencyModeMain,
		},
		LocalMediaInfo: &csta.LocalMediaInfo{
			RTPAddress: &csta.NetworkEndpoint{
				Address: localRtpEndpoint.IP.String(),
				Port:    localRtpEndpoint.Port,
			},
			RTCPAddress: &csta.NetworkEndpoint{
				Address: localRtpEndpoint.IP.String(),
				Port:    65000,
			},
			Codecs:         []string{"g711U"},
			EncryptionList: []string{"none"},
			PacketSize:     20,
		},
	}, func(c *csta.Context) {
		defer wg.Done()
	})

	wg.Wait()
	return nil
}

type monitorPoint struct {
	crossReferenceID string
	device           *device
	events           chan csta.Message
}

func (mp *monitorPoint) CrossReferenceID() string {
	return mp.crossReferenceID
}

func (mp *monitorPoint) Device() pbx.Device {
	return mp.device
}

func (mp *monitorPoint) Events() <-chan csta.Message {
	if mp.events == nil {
		mp.events = make(chan csta.Message, defaultEventBufferSize)
	}
	return mp.events
}

type device struct {
	extension string
	deviceId  string
}

func (d *device) DeviceID() string {
	return d.deviceId
}
