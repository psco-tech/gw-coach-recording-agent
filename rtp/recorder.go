package rtp

import (
	"context"
	"log"
	"net"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/pd0mz/go-g711"
	"github.com/pion/rtp"
)

const defaultReceiveBufferSize = 4096

type Recorder interface {
	// IsRecording returns true if the Recorder is currently recording to a file
	IsRecording() bool

	// StartRecording tries to create the file at filePath and start recording there
	StartRecording(writer *os.File) error

	// StopRecording stops an ongoing recording, it's a no-op if the Recorder is currently idle
	StopRecording() error

	LocalAddr() net.Addr

	// Start starts listening for data and background processing
	Start()
}

type rtpRecorder struct {
	conn    net.PacketConn
	ctx     context.Context
	record  bool
	encoder *wav.Encoder
	file    *os.File
}

// StartRecording starts the recording on this receiver to the filePath specified
func (r *rtpRecorder) StartRecording(writer *os.File) error {
	r.encoder = wav.NewEncoder(writer, 8000, 16, 1, 1)
	r.file = writer
	r.record = true
	return nil
}

// StopRecording stops the recording on this receiver and closes the file
func (r *rtpRecorder) StopRecording() error {
	r.record = false
	r.encoder.Close()
	r.file.Close()
	return nil
}

// Receive handles incoming RTP packets
func (r *rtpRecorder) Start() {
	log.Printf("Listening for incoming RTP data at %s\n", r.conn.LocalAddr().String())

	buf := make([]byte, defaultReceiveBufferSize)
	packet := rtp.Packet{}

	// TODO implement receiving data
	for {
		select {
		case <-r.ctx.Done():
			log.Printf("Shutting down RTP recorder at %s\n", r.conn.LocalAddr())
			r.conn.Close()
			return
		default:
		}

		if r.record {
			n, remote, err := r.conn.ReadFrom(buf)
			if err != nil {
				log.Printf("Failed to read from RTP stream: %s\n", err)
				return
			}

			err = packet.Unmarshal(buf[0:n])
			if err != nil {
				log.Printf("Failed to unmarshal packet from %s: %s\n", remote.String(), err)
				continue
			}

			log.Printf("%+v\n", packet)

			switch packet.PayloadType {
			case 0: // PCMU
				err = r.encoder.Write(toIntBuffer(g711.MLawDecode(packet.Payload)))
				if err != nil {
					log.Printf("Failed to write payload: %s\n", err)
				}
			case 8: // PCMA
				err = r.encoder.Write(toIntBuffer(g711.MLawDecode(packet.Payload)))
				if err != nil {
					log.Printf("Failed to write payload: %s\n", err)
				}
			default:
				log.Printf("Unhandled RTP Payload Type: %d\n", packet.PayloadType)
			}
		}
	}
}

func (r *rtpRecorder) IsRecording() bool {
	return r.record
}

func (r *rtpRecorder) LocalAddr() net.Addr {
	return r.conn.LocalAddr()
}

func toIntBuffer(payload []int16) *audio.IntBuffer {
	buffer := &audio.IntBuffer{
		Data:           make([]int, len(payload)),
		Format:         &audio.Format{NumChannels: 1, SampleRate: 8000},
		SourceBitDepth: 16,
	}

	for i, sample := range payload {
		buffer.Data[i] = int(sample)
	}

	return buffer
}
