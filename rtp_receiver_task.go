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

func (r *RTPReceiverTask) Start(portCount uint) error {
	log.Printf("Starting RTP Receiver Task\n")

	log.Printf("Starting %d RTP listeners\n", portCount)
	r.receivers = make([]*rtpReceiver, portCount)
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

func (r *RTPReceiverTask) GetReceiver() *rtpReceiver {
	log.Printf("Looking for idle receiver\n")
	for i, r := range r.receivers {
		if !r.record {
			log.Printf("Receiver %d is idle\n", i)
			return r
		}
	}
	log.Printf("No idle receiver found!\n")
	return nil
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
