package passive_monitoring

import (
	"context"
	"fmt"
	"log"
	"math"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
	"github.com/google/gopacket"
	"github.com/google/gopacket/ip4defrag"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/pd0mz/go-g711"
	"github.com/pion/rtp"
	"github.com/pion/sdp"
	"github.com/psco-tech/gw-coach-recording-agent/models"
	"github.com/psco-tech/gw-coach-recording-agent/uploader"
)

type Recorder interface {
	ListenAndRecord(ctx context.Context) error // Start listening for calls and record them
}

func NewPassiveRecorder(handle *pcap.Handle) (Recorder, error) {
	return &passiveRecorder{
		handle: handle,
		calls:  make(map[string]*sipCall),
	}, nil
}

type passiveRecorder struct {
	handle *pcap.Handle
	calls  map[string]*sipCall
}

func (r *passiveRecorder) ListenAndRecord(ctx context.Context) error {
	packetSource := gopacket.NewPacketSource(r.handle, r.handle.LinkType())
	defragger := ip4defrag.NewIPv4Defragmenter()

	for packet := range packetSource.Packets() {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		ip4Layer := packet.Layer(layers.LayerTypeIPv4)
		if ip4Layer == nil {
			continue
		}

		ip4 := ip4Layer.(*layers.IPv4)
		defragmentedIPv4, err := defragger.DefragIPv4(ip4)
		if err != nil {
			log.Printf("ERROR: Failed to defragment IPv4 packet: %s\n", err)
			continue
		} else if defragmentedIPv4 == nil {
			// We do not have a full packet yet
			continue
		}

		decoder := defragmentedIPv4.NextLayerType()
		pb, ok := packet.(gopacket.PacketBuilder)
		if !ok {
			log.Printf("ERROR: not a packet builder\n")
			continue
		}

		decoder.Decode(defragmentedIPv4.Payload, pb)

		if sipLayer := packet.Layer(layers.LayerTypeSIP); sipLayer != nil {
			// SIP packet
			sip := sipLayer.(*layers.SIP)
			switch sip.Method {
			case layers.SIPMethodInvite:
				// Create a new call for each invite
				if sip.GetFirstHeader("content-type") == "application/sdp" {
					if _, ok := r.calls[sip.GetCallID()]; !ok {
						log.Printf("Call initiated: %s\n", sip.GetCallID())
						r.calls[sip.GetCallID()] = &sipCall{
							Invite: sip,
						}
					}
				}
			case layers.SIPMethodCancel:
				if _, ok := r.calls[sip.GetCallID()]; ok {
					delete(r.calls, sip.GetCallID())
					log.Printf("Call cancelled: %s", sip.GetCallID())
				}
			case layers.SIPMethodBye:
				if call, ok := r.calls[sip.GetCallID()]; ok {
					delete(r.calls, sip.GetCallID())
					log.Printf("\n\nCall cleared: %s\n%+v\n\n", sip.GetCallID(), call)
					call.Recorder.Close()

					// Enqueue uploading the newly recorded file
					go func() {
						uploader.GetUploadRecordChannel() <- models.UploadRecord{
							FilePath:    call.Recorder.File.Name(),
							Type:        models.UploadRecordTypeCFS_AUDIO,
							ContentType: "audio/wav",
						}

						// TODO this is a stub, create metadata file here
						invite := string(call.Invite.Contents)
						log.Printf("Original Invite: %s", invite)
					}()
				}
			}

			if sip.IsResponse && sip.ResponseCode == 200 {
				if sip.GetFirstHeader("content-type") == "application/sdp" {
					// Find the matching initiated call
					if call, ok := r.calls[sip.GetCallID()]; ok {
						log.Printf("Call established: %s\n", sip.GetCallID())
						call.OK = sip

						// Create the flows from the endpoints

						caller := sdp.SessionDescription{}
						err := caller.Unmarshal(string(call.Invite.Payload()))
						if err != nil {
							log.Printf("ERROR parsing SDP: %s\n", err)
							delete(r.calls, sip.GetCallID())
							continue
						}

						callee := sdp.SessionDescription{}
						err = callee.Unmarshal(string(call.OK.Payload()))
						if err != nil {
							log.Printf("ERROR parsing SDP: %s\n", err)
							delete(r.calls, sip.GetCallID())
							continue
						}

						call.ToCallee = rtpFlow{
							Endpoint: layers.NewIPEndpoint(callee.ConnectionInformation.Address.IP),
							Port:     layers.UDPPort(callee.MediaDescriptions[0].MediaName.Port.Value),
						}

						call.ToCaller = rtpFlow{
							Endpoint: layers.NewIPEndpoint(caller.ConnectionInformation.Address.IP),
							Port:     layers.UDPPort(caller.MediaDescriptions[0].MediaName.Port.Value),
						}

						recordingFile, err := os.CreateTemp(os.TempDir(), "*.wav")

						if err != nil {
							log.Printf("Failed to open file for recording: %s\n", err)
							delete(r.calls, sip.GetCallID())
							continue
						}

						call.Recorder = newRecorder(recordingFile, 2)
					}
				}
			}
		} else if udpLayer := packet.Layer(layers.LayerTypeUDP); udpLayer != nil {
			// UDP but not SIP
			udp := udpLayer.(*layers.UDP)

			// Check if this belongs to any of the active calls RTP streams
			for _, call := range r.calls {
				if packet.NetworkLayer().NetworkFlow().Dst() == call.ToCaller.Endpoint && udp.DstPort == call.ToCaller.Port {
					err := call.Recorder.recordPacket(packet, 0)
					if err != nil {
						log.Printf("ERROR: %s\n", err)
					}
				} else if packet.NetworkLayer().NetworkFlow().Dst() == call.ToCallee.Endpoint && udp.DstPort == call.ToCallee.Port {
					err := call.Recorder.recordPacket(packet, 1)
					if err != nil {
						log.Printf("ERROR: %s\n", err)
					}
				}
			}
		}
	}

	return nil
}

type rtpFlow struct {
	Endpoint gopacket.Endpoint
	Port     layers.UDPPort
}

type sipCall struct {
	Invite *layers.SIP
	OK     *layers.SIP

	ToCaller rtpFlow
	ToCallee rtpFlow

	Recorder *multichannelRecorder
}

func newRecorder(file *os.File, channels int) *multichannelRecorder {
	recorder := &multichannelRecorder{
		Encoder: wav.NewEncoder(file, 8000, 16, channels, 1),
		File:    file,
		buffers: make([]*audio.IntBuffer, channels),
	}

	for i := 0; i < channels; i++ {
		recorder.buffers[i] = &audio.IntBuffer{
			Data:           make([]int, 0),
			Format:         &audio.Format{NumChannels: 1, SampleRate: 8000},
			SourceBitDepth: 16,
		}
	}

	return recorder
}

type multichannelRecorder struct {
	Encoder *wav.Encoder
	File    *os.File

	buffers []*audio.IntBuffer
}

func (r *multichannelRecorder) recordPacket(packet gopacket.Packet, channel int) error {
	rtp := rtp.Packet{}
	err := rtp.Unmarshal(packet.ApplicationLayer().Payload())
	if err != nil {
		return fmt.Errorf("failed to unmarshal RTP packet: %s", err)
	}

	// Write new samples into according buffers
	switch rtp.PayloadType {
	case 0: // PCMU
		decodedSamples := g711.MLawDecode(rtp.Payload)
		newSamples := make([]int, len(decodedSamples))
		for i := 0; i < len(newSamples); i++ {
			newSamples[i] = int(decodedSamples[i])
		}

		r.buffers[channel].Data = append(r.buffers[channel].Data, newSamples...)
	case 8: // PCMA
		decodedSamples := g711.ALawDecode(rtp.Payload)
		newSamples := make([]int, len(decodedSamples))
		for i := 0; i < len(newSamples); i++ {
			newSamples[i] = int(decodedSamples[i])
		}

		r.buffers[channel].Data = append(r.buffers[channel].Data, newSamples...)
	}

	n, samples := r.interleave(r.buffers)
	for _, b := range r.buffers {
		b.Data = b.Data[n:]
	}
	return r.Encoder.Write(samples)
}

func (r *multichannelRecorder) interleave(channels []*audio.IntBuffer) (int, *audio.IntBuffer) {
	samples := math.MaxInt
	for _, channel := range channels {
		if len(channel.Data) < samples {
			samples = len(channel.Data)
		}
	}

	interleaved := &audio.IntBuffer{
		Data:           make([]int, 0),
		Format:         &audio.Format{NumChannels: len(channels), SampleRate: 8000},
		SourceBitDepth: 16,
	}

	for i := 0; i < samples; i++ {
		for c := 0; c < len(channels); c++ {
			interleaved.Data = append(interleaved.Data, channels[c].Data[i])
		}
	}

	return samples, interleaved
}

func (r *multichannelRecorder) Close() error {
	r.Encoder.Close()
	return r.File.Close()
}
