package avaya

import (
	"context"
	"fmt"
	"io"
	"log"
	"net"
	"sync"
	"time"

	"github.com/psco-tech/gw-coach-recording-agent/csta"
	"github.com/psco-tech/gw-coach-recording-agent/models"
	"github.com/psco-tech/gw-coach-recording-agent/pbx"
	"github.com/psco-tech/gw-coach-recording-agent/rtp"
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
func (aes *AvayaAES) Connect() (csta.Conn, error) {
	cstaConn, err := csta.Dial("tcp", viper.GetString("avaya_aes.server_address"), aes.ctx, nil)
	if err != nil {
		return nil, err
	}

	aes.conn = cstaConn
	aes.setupHandlers()

	var wg sync.WaitGroup

	wg.Add(1)
	err = cstaConn.StartApplicationSession(viper.GetString("application_id"), struct {
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
			Username:            viper.GetString("avaya_aes.username"),
			Password:            viper.GetString("avaya_aes.password"),
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

func (aes *AvayaAES) setupHandlers() {
	aes.conn.Handle(csta.MessageTypeEstablishedEvent, aes.onEstablishedEvent)
	aes.conn.Handle(csta.MessageTypeConnectionClearedEvent, aes.onConnectionClearedEvent)
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

func (aes *AvayaAES) Serve(recorderPool rtp.RecorderPool) error {
	defer aes.Close()
	log.Printf("Handling PBX connection\n")

	// Get access to the persistence layer
	db, err := models.NewDatabase()
	if err != nil {
		return err
	}

	// Get recprding devices to be registered
	recordingDevices := db.GetAESRecordingDevices()
	log.Printf("%d AES recording devices configured\n", len(recordingDevices))
	recorders := recorderPool.GetAllRecorders()

	if len(recorders) < len(recordingDevices) {
		return fmt.Errorf("not enough recorders configured to service all recording devices")
	}

	for i, rd := range recordingDevices {
		log.Printf("Registering AES recording device <%s> with local recording endpoint <%s>", rd.Extension, recorders[i].LocalAddr().String())
		err := aes.RegisterTerminal(rd.Extension, rd.Password, recorders[i].LocalAddr().(*net.UDPAddr))
		if err != nil {
			log.Printf("Failed to register AES recording device: %s\n", err)
		}
	}

	// Get devices to be monitored
	monitoredDevices := db.GetMonitoredDevices()
	log.Printf("%d devices are configured to be monitored\n", len(monitoredDevices))

	// Run MonitorStart on all of the monitored devices
	for _, d := range monitoredDevices {
		mp, err := aes.MonitorStart(d.Extension)
		if err != nil {
			log.Printf("Failed to start monitoring <%s>: %s", d.Extension, err)
			continue
		}

		d.CrossReferenceID = mp.CrossReferenceID()
		db.Save(d)
	}

	// Add additional actions to do on newly established PBX connection here

	// Handlers will run in the background, wait for anything to fail/end
	select {
	case <-aes.ctx.Done():
		return nil
	case <-aes.conn.Closed():
		return io.EOF
	}
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
	subscribers      []chan csta.Message
}

func (mp *monitorPoint) CrossReferenceID() string {
	return mp.crossReferenceID
}

func (mp *monitorPoint) Device() pbx.Device {
	return mp.device
}

func (mp *monitorPoint) Events() <-chan csta.Message {
	if mp.subscribers == nil {
		mp.subscribers = make([]chan csta.Message, 0)
	}
	subscriberChannel := make(chan csta.Message, defaultEventBufferSize)
	mp.subscribers = append(mp.subscribers, subscriberChannel)
	return subscriberChannel
}

func (mp *monitorPoint) dispatchEvent(e csta.Message) {
	for _, subscriber := range mp.subscribers {
		subscriber <- e
	}
}

type device struct {
	extension string
	deviceId  string
}

func (d *device) DeviceID() string {
	return d.deviceId
}

func (aes *AvayaAES) onEstablishedEvent(c *csta.Context) {
	// Check for the correct event data type
	if event, ok := (c.Message).(*csta.EstablishedEvent); ok {
		// Get the monitor point this event is for
		if mp, ok := aes.monitorPoints[event.MonitorCrossRefID]; ok {
			mp.dispatchEvent(event)
		}
	}
}

func (aes *AvayaAES) onConnectionClearedEvent(c *csta.Context) {
	// Check for the correct event data type
	if event, ok := (c.Message).(*csta.ConnectionClearedEvent); ok {
		// Get the monitor point this event is for
		if mp, ok := aes.monitorPoints[event.MonitorCrossRefID]; ok {
			mp.dispatchEvent(event)
		}
	}
}
