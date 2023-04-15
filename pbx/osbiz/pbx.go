package osbiz

import (
	"fmt"
	"log"
	"sync"

	"github.com/psco-tech/gw-coach-recording-agent/csta"
	"github.com/psco-tech/gw-coach-recording-agent/pbx"
)

const defaultEventBufferSize = 100

type OSBiz struct {
	conn          csta.Conn
	monitorPoints map[string]*monitorPoint
}

func (pbx *OSBiz) Connect(network, addr, applicationId, username, password string) error {
	cstaConn, err := csta.Dial(network, addr, nil)
	if err != nil {
		return err
	}

	pbx.setupHandlers(cstaConn)

	var wg sync.WaitGroup

	pbx.conn = cstaConn

	wg.Add(1)
	err = cstaConn.StartApplicationSession(applicationId, struct {
		User     string `xml:"user"`
		Password string `xml:"password"`
	}{
		User:     username,
		Password: password,
	}, func(ctx *csta.Context) {
		if ctx.Error != nil {
			err = ctx.Error
		}
		wg.Done()
	})

	if err != nil {
		return err
	}

	wg.Wait()

	return nil
}

func (pbx *OSBiz) getMonitorPoint(crossReferenceId string) *monitorPoint {
	if mp, ok := pbx.monitorPoints[crossReferenceId]; ok {
		return mp
	}
	return nil
}

func (o *OSBiz) setupHandlers(conn csta.Conn) {
	conn.Handle(csta.MessageTypeEstablishedEvent, func(c *csta.Context) {
		if e, ok := (c.Message).(*csta.EstablishedEvent); ok {
			mp := o.getMonitorPoint(e.MonitorCrossRefID)
			mp.dispatchEvent(e)
		}
	})

	conn.Handle(csta.MessageTypeOutOfServiceEvent, func(c *csta.Context) {
		if e, ok := (c.Message).(*csta.OutOfServiceEvent); ok {
			mp := o.getMonitorPoint(e.MonitorCrossRefID)
			mp.dispatchEvent(e)
		}
	})

	conn.Handle(csta.MessageTypeBackInServiceEvent, func(c *csta.Context) {
		if e, ok := (c.Message).(*csta.BackInServiceEvent); ok {
			mp := o.getMonitorPoint(e.MonitorCrossRefID)
			mp.dispatchEvent(e)
		}
	})
}

func (osbiz *OSBiz) ConnectionState() pbx.ConnectionState {
	switch osbiz.conn.State() {
	case csta.ConnectionStateActive:
		return pbx.ConnectionStateConnected
	case csta.ConnectionStateError:
		return pbx.ConnectionStateError
	}

	return pbx.ConnectionStateDisconnected
}

func (pbx *OSBiz) Close() error {
	return pbx.conn.Close()
}

func (osbiz *OSBiz) MonitorStart(deviceId string) (mp pbx.MonitorPoint, err error) {
	log.Printf("MonitorStart(<%s>)", deviceId)

	var wg sync.WaitGroup
	wg.Add(1)

	err = osbiz.conn.MonitorStart(csta.CSTAObject{
		DeviceObject: &csta.DeviceID{Device: deviceId, TypeOfNumber: "dialingNumber"},
	}, csta.MonitorTypeDevice, func(c *csta.Context) {
		defer wg.Done()

		if c.Error != nil {
			err = c.Error
			return
		}

		if resp, ok := (c.Message).(*csta.MonitorStartResponse); ok {
			if osbiz.monitorPoints == nil {
				osbiz.monitorPoints = make(map[string]*monitorPoint)
			}
			osbiz.monitorPoints[resp.MonitorCrossRefID] = &monitorPoint{
				crossReferenceID: resp.MonitorCrossRefID,
				device: &device{
					extension: deviceId,
				},
			}
			mp = osbiz.monitorPoints[resp.MonitorCrossRefID]
			return
		}

		err = fmt.Errorf("failed to monitor device, unknown error")
	})

	wg.Wait()

	if err == nil {
		log.Printf("Monitoring <%s> with CrossRefID <%s>\n", deviceId, mp.CrossReferenceID())
	}

	return
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

func (mp *monitorPoint) dispatchEvent(e csta.Message) {
	if mp.events == nil {
		mp.events = make(chan csta.Message, defaultEventBufferSize)
	}
	mp.events <- e
}

type device struct {
	extension string
}

func (d *device) DeviceID() string {
	return d.extension
}
