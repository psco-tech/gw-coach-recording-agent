package rtp

import (
	"context"
	"log"
	"net"
)

type Recorder interface {
	// IsRecording returns true if the Recorder is currently recording to a file
	IsRecording() bool

	// StartRecording tries to create the file at filePath and start recording there
	StartRecording(filePath string) error

	// StopRecording stops an ongoing recording, it's a no-op if the Recorder is currently idle
	StopRecording() error

	LocalAddr() net.Addr

	// Start starts listening for data and background processing
	Start()
}

type rtpRecorder struct {
	conn   net.PacketConn
	ctx    context.Context
	record bool
}

// StartRecording starts the recording on this receiver to the filePath specified
func (r *rtpRecorder) StartRecording(filePath string) error {
	r.record = true
	return nil
}

// StopRecording stops the recording on this receiver and closes the file
func (r *rtpRecorder) StopRecording() error {
	r.record = false
	return nil
}

// Receive handles incoming RTP packets
func (r *rtpRecorder) Start() {
	log.Printf("Listening for incoming RTP data at %s\n", r.conn.LocalAddr().String())

	// TODO implement receiving data

	select {
	case <-r.ctx.Done():
		log.Printf("Shutting down RTP recorder at %s\n", r.conn.LocalAddr())
		r.conn.Close()
		return
	}
}

func (r *rtpRecorder) IsRecording() bool {
	return r.record
}

func (r *rtpRecorder) LocalAddr() net.Addr {
	return r.conn.LocalAddr()
}
