package recorder

import (
	"bytes"
	"errors"
	"log"
	"os"
	"os/signal"
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

	VADConfig *vad.VADConfig
}

type Recorder struct {
	cfg    *RecorderConfig
	stream *stream.Stream
	vad    *vad.VAD

	quit chan bool
}

func DefaultRecorderConfig() *RecorderConfig {
	return &RecorderConfig{
		SampleRate:      22050,
		InputChannels:   1,
		FramesPerBuffer: 64,
		MaxTime:         100000,

		VADConfig: vad.DefaultVADConfig(),
	}
}

func NewRecorder(cfg *RecorderConfig, stream *stream.Stream) (*Recorder, error) {
	if cfg == nil {
		return nil, ErrInvalidRecorderConfig
	}

	vad := vad.NewVAD(cfg.VADConfig)
	return &Recorder{
		cfg:    cfg,
		stream: stream,
		vad:    vad,

		quit: make(chan bool),
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
	signalCh := make(chan os.Signal, 1)
	signal.Notify(signalCh, os.Interrupt)

	err := r.stream.Start()
	if err != nil {
		return nil, err
	}
	defer r.stream.Close()

	go func() {
		detectStop := make(chan bool)
		for {
			if r.vad.DetectSpeech(detectStop) {
				break
			}

			quit := false
			select {
			case <-signalCh:
				r.quit <- true
				detectStop <- true
				quit = true
			default:
			}

			if quit {
				break
			}
		}

		speechCh <- true
	}()

	for {
		buffer, err := r.stream.Read()
		if err != nil {
			log.Printf("stream error -- %s", err)
			return nil, err
		}
		r.vad.AddBuffer(buffer)

		quit := false
		select {
		case <-speechCh:
			quit = true
		case <-r.quit:
			quit = true
		default:
			continue
		}

		if quit {
			break
		}
	}

	log.Printf("Waiting...")
	go func() {
		detectStop := make(chan bool)
		for {
			if r.vad.DetectSilence(detectStop) {
				break
			}

			quit := false
			select {
			case <-signalCh:
				r.quit <- true
				detectStop <- true
				quit = true
			default:
			}
			if quit {
				break
			}
		}

		speechCh <- true
	}()

	fullStream := []int32{}
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
		case <-r.quit:
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
