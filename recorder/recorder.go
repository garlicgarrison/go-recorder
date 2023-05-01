package recorder

import (
	"bytes"
	"errors"
	"log"
	"time"

	"github.com/garlicgarrison/go-recorder/codec"
	"github.com/garlicgarrison/go-recorder/stream"
	"github.com/garlicgarrison/go-recorder/vad"
)

type Format string

const (
	WAV  Format = "wav"
	AIFF Format = "aiff"
)

var (
	ErrInvalidRecorderConfig = errors.New("invalid config")
)

type RecorderConfig struct {
	SampleRate      float64
	InputChannels   int
	FramesPerBuffer int
	MaxTime         int //milliseconds
}

type Recorder struct {
	cfg    *RecorderConfig
	stream *stream.Stream
	vad    *vad.VAD
}

func DefaultRecorderConfig() *RecorderConfig {
	return &RecorderConfig{
		SampleRate:      22050,
		InputChannels:   1,
		FramesPerBuffer: 64,
		MaxTime:         100000,
	}
}

func NewRecorder(cfg *RecorderConfig, stream *stream.Stream) (*Recorder, error) {
	if cfg == nil {
		return nil, ErrInvalidRecorderConfig
	}

	vad := vad.NewVAD(vad.DefaultVADConfig())
	return &Recorder{
		cfg:    cfg,
		stream: stream,
		vad:    vad,
	}, nil
}

func (r *Recorder) Record(format Format, quit chan bool) (*bytes.Buffer, error) {
	timerChan := make(chan bool)
	go func() {
		time.Sleep(time.Millisecond * time.Duration(r.cfg.MaxTime))
		timerChan <- true
	}()

	fullStream := []int32{}
	for {
		buffer, err := r.stream.Read()
		if err != nil {
			return nil, err
		}

		currStream := make([]int32, r.cfg.FramesPerBuffer)
		copy(currStream, buffer)
		fullStream = append(fullStream, currStream...)
		select {
		case <-quit:
			switch format {
			case AIFF:
				return codec.NewDefaultAIFF(fullStream).EncodeAIFF()
			case WAV:
				return codec.NewDefaultWAV(fullStream).EncodeWAV()
			}
		case <-timerChan:
			switch format {
			case AIFF:
				return codec.NewDefaultAIFF(fullStream).EncodeAIFF()
			case WAV:
				return codec.NewDefaultWAV(fullStream).EncodeWAV()
			}
		default:
		}
	}
}

func (r *Recorder) RecordVAD(format Format) (*bytes.Buffer, error) {
	log.Printf("Listening...")
	speechCh := make(chan bool)
	go func() {
		for {
			if r.vad.DetectSpeech() {
				break
			}
		}
		speechCh <- true
	}()

	for {
		buffer, err := r.stream.Read()
		if err != nil {
			return nil, err
		}
		r.vad.AddBuffer(buffer)

		quit := false
		select {
		case <-speechCh:
			quit = true
		default:
			continue
		}

		if quit {
			break
		}
	}

	fullStream := []int32{}
	go func() {
		for {
			if r.vad.DetectSilence() {
				break
			}
		}

		speechCh <- true
	}()

	for {
		buffer, err := r.stream.Read()
		if err != nil {
			return nil, err
		}
		r.vad.AddBuffer(buffer)
		fullStream = append(fullStream, buffer...)

		quit := false
		select {
		case <-speechCh:
			quit = true
		default:
			continue
		}

		if quit {
			break
		}
	}

	log.Printf("Stopped...")
	switch format {
	case AIFF:
		return codec.NewDefaultAIFF(fullStream).EncodeAIFF()
	case WAV:
		return codec.NewDefaultWAV(fullStream).EncodeWAV()
	default:
		return codec.NewDefaultWAV(fullStream).EncodeWAV()
	}
}
