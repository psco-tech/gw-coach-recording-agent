package osbiz

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/psco-tech/gw-coach-recording-agent/csta"
	"github.com/psco-tech/gw-coach-recording-agent/pbx"
	"github.com/spf13/viper"
)

const defaultEventBufferSize = 100

type OSBiz struct {
	ctx           context.Context
	sessionId     string
	conn          csta.Conn
	monitorPoints map[string]*monitorPoint
}

func (pbx *OSBiz) SetContext(ctx context.Context) {
	pbx.ctx = ctx
}

func (pbx *OSBiz) Serve() error {
	defer pbx.Close()
	return nil
}

func (pbx *OSBiz) Connect() (csta.Conn, error) {
	cstaConn, err := csta.Dial("tcp", viper.GetString("osbiz.server_address"), pbx.ctx, nil)
	if err != nil {
		return nil, err
	}

	pbx.setupHandlers(cstaConn)

	var wg sync.WaitGroup

	pbx.conn = cstaConn

	wg.Add(1)
	err = cstaConn.StartApplicationSession(viper.GetString("application_id"), struct {
		User     string `xml:"user"`
		Password string `xml:"password"`
	}{
		User:     viper.GetString("osbiz.username"),
		Password: viper.GetString("osbiz.password"),
	}, "http://www.ecma-international.org/standards/ecma-323/csta/ed4",
		func(ctx *csta.Context) {
			if ctx.Error != nil {
				err = ctx.Error
			}
			if r, ok := ctx.Message.(*csta.StartApplicationSessionPosResponse); ok {
				log.Printf("Application session started with session id <%s>\n", r.SessionID)
				pbx.sessionId = r.SessionID
			}
			wg.Done()
		})

	if err != nil {
		return nil, err
	}

	wg.Wait()

	return cstaConn, nil
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
	var wg sync.WaitGroup
	wg.Add(1)
	pbx.conn.Request(csta.StopApplicationSession{
		SessionID:        pbx.sessionId,
		SessionEndReason: "Application Shutdown",
	}, func(c *csta.Context) {
		wg.Done()
	})
	wg.Wait()

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
