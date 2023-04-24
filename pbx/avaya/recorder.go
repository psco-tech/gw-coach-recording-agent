package avaya

import (
	"fmt"
	"os"

	"github.com/psco-tech/gw-coach-recording-agent/rtp"
)

type recorderTerminal struct {
	Extension   string
	CurrentCall string
	Recorder    rtp.Recorder
}

func (aes *AvayaAES) GetRecorder() (*recorderTerminal, error) {
	for _, r := range aes.recorders {
		if !r.Recorder.IsRecording() {
			return r, nil
		}
	}
	return nil, fmt.Errorf("no idle RTP receiver available")
}

func (aes *AvayaAES) GetRecorderByCallReference(callReference string) (*recorderTerminal, error) {
	for _, r := range aes.recorders {
		if r.CurrentCall == callReference {
			return r, nil
		}
	}
	return nil, fmt.Errorf("no recorder for call reference <%s>", callReference)
}

func (r *recorderTerminal) StartRecording(writer *os.File, callReference string) error {
	r.CurrentCall = callReference
	return r.Recorder.StartRecording(writer)
}

func (r *recorderTerminal) StopRecording() error {
	r.CurrentCall = ""
	return r.Recorder.StopRecording()
}
