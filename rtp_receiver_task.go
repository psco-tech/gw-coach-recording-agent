package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"sync"

	"github.com/spf13/viper"
)

func init() {
	viper.SetDefault("rtp_receiver_count", 16)
	viper.SetDefault("rtp_receiver_addr", "0.0.0.0")
}

type rtpReceiver struct {
	conn   net.PacketConn
	wg     *sync.WaitGroup
	record bool
}

// The RTPReceiverTask handles everythin related to receiving, decoding and recording RTP packets
type RTPReceiverTask struct {
	ctx       context.Context
	wg        *sync.WaitGroup
	receivers []*rtpReceiver
}

func NewRTPReceiverTask(ctx context.Context, wg *sync.WaitGroup) *RTPReceiverTask {
	return &RTPReceiverTask{
		ctx: ctx,
		wg:  wg,
	}
}

// Start will start the RTPReceiverTasks background operation
// portCount sets the number of RTP receivers to instantiate
func (r *RTPReceiverTask) Start(portCount uint) error {
	log.Printf("Starting RTP Receiver Task\n")

	log.Printf("Starting %d RTP listeners\n", portCount)
	r.receivers = make([]*rtpReceiver, 0)
	for i := 0; i < int(portCount); i++ {
		l, err := net.ListenPacket("udp4", fmt.Sprintf("%s:0", viper.GetString("rtp_receiver_addr")))
		if err != nil {
			log.Printf("Failed to start RTP listener: %s", err)
			continue
		}

		receiver := rtpReceiver{conn: l, wg: r.wg}
		r.receivers = append(r.receivers, &receiver)
		go receiver.receive()
	}

	return nil
}

func (r *RTPReceiverTask) GetRecorder() (Recorder, error) {
	for _, r := range r.receivers {
		if !r.record {
			return r, nil
		}
	}
	return nil, fmt.Errorf("no idle RTP receiver available")
}

func (r *RTPReceiverTask) GetAllRecorders() (recorders []Recorder) {
	recorders = make([]Recorder, len(r.receivers))
	for i, r := range r.receivers {
		recorders[i] = r
	}
	return
}

func (r *rtpReceiver) StartRecording(filePath string) error {
	r.record = true
	return nil
}

func (r *rtpReceiver) StopRecording() error {
	r.record = false
	return nil
}

func (r *rtpReceiver) receive() {
	log.Printf("Listening for incoming RTP data at %s\n", r.conn.LocalAddr().String())

	r.wg.Add(1)
	defer r.wg.Done()

}

func (r *rtpReceiver) IsRecording() bool {
	return r.record
}

func (r *rtpReceiver) LocalAddr() net.Addr {
	return r.conn.LocalAddr()
}
