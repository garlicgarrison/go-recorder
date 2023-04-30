package recorder

import (
	"bytes"
	"errors"
	"os"
	"time"

	"github.com/garlicgarrison/go-recorder/codec"
	"github.com/gordonklaus/portaudio"
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
	buffer []int32
}

func NewDefaultRecorderConfig() *RecorderConfig {
	return &RecorderConfig{
		SampleRate:      22050,
		InputChannels:   1,
		FramesPerBuffer: 64,
		MaxTime:         100000,
	}
}

func NewRecorder(cfg *RecorderConfig) (*Recorder, error) {
	if cfg == nil {
		return nil, ErrInvalidRecorderConfig
	}

	return &Recorder{
		cfg:    cfg,
		buffer: make([]int32, cfg.FramesPerBuffer),
	}, nil
}

func (r *Recorder) Record(format Format, quit chan os.Signal) (*bytes.Buffer, error) {
	timerChan := make(chan bool)
	go func() {
		time.Sleep(time.Millisecond * time.Duration(r.cfg.MaxTime))
		timerChan <- true
	}()

	err := portaudio.Initialize()
	if err != nil {
		return nil, err
	}

	defer portaudio.Terminate()

	stream, err := portaudio.OpenDefaultStream(
		r.cfg.InputChannels,
		0,
		r.cfg.SampleRate,
		r.cfg.FramesPerBuffer,
		r.buffer,
	)
	if err != nil {
		return nil, err
	}

	err = stream.Start()
	if err != nil {
		return nil, err
	}

	defer stream.Close()

	fullStream := []int32{}
	for {
		err = stream.Read()
		if err != nil {
			return nil, err
		}

		currStream := make([]int32, r.cfg.FramesPerBuffer)
		copy(currStream, r.buffer)
		fullStream = append(fullStream, currStream...)
		select {
		case <-quit:
			if err = stream.Stop(); err != nil {
				return nil, err
			}

			switch format {
			case AIFF:
				return codec.NewDefaultAIFF(fullStream).EncodeAIFF()
			case WAV:
				return codec.NewDefaultWAV(fullStream).EncodeWAV()
			}
		case <-timerChan:
			if err = stream.Stop(); err != nil {
				return nil, err
			}

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
