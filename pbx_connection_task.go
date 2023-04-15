package main

import (
	"context"
	"errors"
	"io"
	"log"
	"sync"
	"time"

	"github.com/psco-tech/gw-coach-recording-agent/csta"
	"github.com/psco-tech/gw-coach-recording-agent/models"
	"github.com/psco-tech/gw-coach-recording-agent/pbx"
	"github.com/psco-tech/gw-coach-recording-agent/pbx/avaya"
	"github.com/psco-tech/gw-coach-recording-agent/pbx/osbiz"
	"github.com/spf13/viper"
)

type Recorder interface {
	// IsRecording returns true if the Recorder is currently recording to a file
	IsRecording() bool

	// StartRecording tries to create the file at filePath and start recording there
	StartRecording(filePath string) error

	// StopRecording stops an ongoing recording, it's a no-op if the Recorder is currently idle
	StopRecording() error
}

// RecorderPool provides a set of recorders to the system to use
type RecorderPool interface {
	// GetRecorder tries to get an available (not currently recording)
	// recorder for the system to use on a new recording session
	GetRecorder() (Recorder, error)

	// GetAllRecorders gets a list of all Recorders that are configured
	GetAllRecorders() []Recorder
}

// The PBXConnectionTask handles everything associated with keeping the connection to the PBX
// up and running as well as everything related to PBX communications and event handling
type PBXConnectionTask struct {
	ctx       context.Context
	sf        pbx.PBX
	wg        *sync.WaitGroup
	recorders RecorderPool
}

func NewPBXConnectionTask(ctx context.Context, wg *sync.WaitGroup) *PBXConnectionTask {
	return &PBXConnectionTask{
		ctx: ctx,
		sf:  nil,
		wg:  wg,
	}
}

func (p *PBXConnectionTask) SetRecorderPool(r RecorderPool) {
	p.recorders = r
}

// Start will start the PBXConnectionTasks background operation
// pbxType will set the vendor specific implementation to use for PBX communications
func (p *PBXConnectionTask) Start(pbxType string) error {
	log.Printf("Starting PBX Connection Task\n")
	pbx.RegisterImplementation("osbiz", &osbiz.OSBiz{})
	pbx.RegisterImplementation("avaya_aes", &avaya.AvayaAES{})

	sf, err := pbx.New(pbxType, p.ctx)
	if err != nil {
		return err
	}

	p.sf = sf

	go p.connectionHandler()

	return nil
}

// The connectionHandler handles establishing a connection and session with the PBX
// If the connection is lost it will retry, if the application is shutting down it will exit
func (p *PBXConnectionTask) connectionHandler() {
	defer p.wg.Done()
	defer log.Printf("Stopped PBX connection handler\n")

	p.wg.Add(1)

	const connectionRetryTimeout = 30 * time.Second
	log.Printf("Starting PBX connection handler\n")

	for {
		select {
		case <-p.ctx.Done():
			log.Printf("Closing PBX Session\n")
			p.sf.Close()
			return
		default:
			log.Printf("Trying to connect to PBX (%s)\n", viper.GetString("pbx_address"))
			cstaConn, err := p.sf.Connect(
				"tcp",
				viper.GetString("pbx_address"),
				viper.GetString("application_id"),
				viper.GetString("pbx_username"),
				viper.GetString("pbx_password"),
			)

			if err != nil {
				log.Printf("Failed to connect to PBX: %s\n", err)

				time.Sleep(connectionRetryTimeout)
				continue
			}

			log.Printf("Successfully connected to PBX\n")

			err = p.pbxHandler(cstaConn)
			if err != nil {
				if !errors.Is(err, io.EOF) {
					log.Printf("PBX error: %s\n", err)
				} else {
					log.Printf("PBX connection closed: %s\n", err)
				}
				log.Printf("Connection retry in %ds\n", connectionRetryTimeout/time.Second)
				// TODO handle if shutdown of application was requested during connectionRetryTimout
				time.Sleep(connectionRetryTimeout)
				continue
			}

			if p.sf != nil {
				log.Printf("Closing PBX Session\n")
				p.sf.Close()
			}

			return
		}
	}
}

// The pbxHandler is the "main loop" that works on handling events and sending messages to the PBX
func (p *PBXConnectionTask) pbxHandler(cstaConn csta.Conn) error {
	defer log.Printf("Stopped PBX handler\n")

	// Get access to the persistence layer
	db, err := models.NewDatabase()
	if err != nil {
		return err
	}

	// Get devices to be monitored on startup
	monitoredDevices := db.GetMonitoredDevices()
	log.Printf("%d devices are configured to be monitored\n", len(monitoredDevices))

	// TODO register recorders with the PBX

	// Start monitoring the devices, save the cross reference ID to the DB
	for _, d := range monitoredDevices {
		mp, err := p.sf.MonitorStart(d.Extension)
		if err != nil {
			log.Printf("Failed to start monitoring <%s>: %s", d.Extension, err)
			continue
		}

		d.CrossReferenceID = mp.CrossReferenceID()
		db.Save(d)
	}

	select {
	case <-p.ctx.Done():
		return nil
	case <-cstaConn.Closed():
		return io.EOF
	}
}
