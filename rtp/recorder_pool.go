package rtp

import (
	"context"
	"fmt"
	"log"
	"net"

	"github.com/spf13/viper"
)

type RecorderPool interface {
	// GetRecorder tries to get an available (not currently recording)
	// recorder for the system to use on a new recording session
	GetRecorder() (Recorder, error)

	// GetAllRecorders gets a list of all Recorders that are configured
	GetAllRecorders() []Recorder

	// Start background processing (handle packets, record to files etc.)
	Start() error
}

func init() {
	viper.SetDefault("rtp.recorder_count", 16)
}

type rtpRecorderPool struct {
	ctx       context.Context
	recorders []*rtpRecorder
}

func NewRecorderPool(size uint, ctx context.Context) (RecorderPool, error) {
	log.Printf("Creating RTP recorder pool with %d ports\n", size)
	pool := &rtpRecorderPool{
		ctx:       ctx,
		recorders: make([]*rtpRecorder, size),
	}
	return pool, nil
}

// Start will start the RTPReceiverTasks background operation
// portCount sets the number of RTP receivers to instantiate
func (r *rtpRecorderPool) Start() error {
	log.Printf("Starting RTP Receiver Task\n")

	log.Printf("Starting %d RTP listeners\n", len(r.recorders))
	for i := 0; i < len(r.recorders); i++ {
		l, err := net.ListenPacket("udp4", fmt.Sprintf("%s:0", viper.GetString("rtp.recorder_address")))
		if err != nil {
			log.Printf("Failed to start RTP listener: %s", err)
			continue
		}

		r.recorders[i] = &rtpRecorder{conn: l, ctx: r.ctx}

		// Start them immediately
		// TODO handle recorder failures
		go r.recorders[i].Start()
	}

	return nil
}

// GetRecorder gets the first recorder that isn't recording at the moment or returns an error if
// the RecorderPool is exhausted
func (r *rtpRecorderPool) GetRecorder() (Recorder, error) {
	for _, r := range r.recorders {
		if !r.record {
			return r, nil
		}
	}
	return nil, fmt.Errorf("no idle RTP receiver available")
}

// GetAllRecorders returns a slice of all configured recorders
func (r *rtpRecorderPool) GetAllRecorders() (recorders []Recorder) {
	recorders = make([]Recorder, len(r.recorders))
	for i, r := range r.recorders {
		recorders[i] = r
	}
	return
}
