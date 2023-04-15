package main

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/psco-tech/gw-coach-recording-agent/models"
	"github.com/psco-tech/gw-coach-recording-agent/pbx"
	"github.com/psco-tech/gw-coach-recording-agent/pbx/avaya"
	"github.com/psco-tech/gw-coach-recording-agent/pbx/osbiz"
	"github.com/spf13/viper"
)

type PBXConnectionTask struct {
	ctx context.Context
	sf  pbx.PBX
	wg  *sync.WaitGroup
}

func NewPBXConnectionTask(ctx context.Context, wg *sync.WaitGroup) *PBXConnectionTask {
	return &PBXConnectionTask{
		ctx: ctx,
		sf:  nil,
		wg:  wg,
	}
}

func (p *PBXConnectionTask) Start(pbxType string) error {
	log.Printf("Starting PBX Connection Task\n")
	pbx.RegisterImplementation("osbiz", &osbiz.OSBiz{})
	pbx.RegisterImplementation("avaya_aes", &avaya.AvayaAES{})

	sf, err := pbx.New(pbxType)
	if err != nil {
		return err
	}

	p.sf = sf

	go p.connectionHandler()

	return nil
}

func (p *PBXConnectionTask) connectionHandler() {
	defer p.wg.Done()
	defer log.Printf("Stopped PBX connection handler\n")

	p.wg.Add(1)

	const connectionRetryTimeout = 30 * time.Second
	log.Printf("Starting PBX connection handler\n")

	for {
		select {
		case <-p.ctx.Done():
			p.sf.Close()
			return
		default:
			log.Printf("Trying to connect to PBX (%s)\n", viper.GetString("pbx_address"))
			err := p.sf.Connect(
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

			err = p.pbxHandler()
			if err != nil {
				log.Printf("PBX error: %s\n", err)

				time.Sleep(connectionRetryTimeout)
				continue
			}

			return
		}
	}
}

func (p *PBXConnectionTask) pbxHandler() error {
	defer log.Printf("Stopped PBX handler\n")

	// Get access to the persistence layer
	db, err := models.NewDatabase()
	if err != nil {
		return err
	}

	// Get devices to be monitored on startup
	monitoredDevices := db.GetMonitoredDevices()
	log.Printf("%d devices are configured to be monitored\n", len(monitoredDevices))

	// Start monitoring the devices, save the cross reference ID to the DB
	for _, d := range monitoredDevices {
		mp, err := p.sf.MonitorStart(d.DeviceID)
		if err != nil {
			log.Printf("Failed to start monitoring <%s>: %s", d.DeviceID, err)
			continue
		}

		d.CrossReferenceID = mp.CrossReferenceID()
		db.Save(d)
	}

	<-p.ctx.Done()
	return nil
}
