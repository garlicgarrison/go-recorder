package stream

import (
	"errors"
	"log"
	"sync"

	"github.com/gordonklaus/portaudio"
)

const (
	DefaultInputChannels   = 1
	DefaultSampleRate      = 22050
	DefaultFramesPerBuffer = 64
)

type StreamState string

const (
	Closed  = "closed"
	Started = "started"
	Opened  = "opened"
)

var (
	ErrAlreadyStarted = errors.New("stream already started")
	ErrAlreadyOpened  = errors.New("stream already opened")
)

type StreamConfig struct {
	SampleRate      float64
	InputChannels   int
	FramesPerBuffer int
}

// Singleton
type Stream struct {
	cfg    *StreamConfig
	stream *portaudio.Stream
	mutex  *sync.Mutex
	state  StreamState
	buffer []int32
}

func DefaultStreamConfig() *StreamConfig {
	return &StreamConfig{
		SampleRate:      DefaultSampleRate,
		InputChannels:   DefaultInputChannels,
		FramesPerBuffer: DefaultFramesPerBuffer,
	}
}

func NewStream(cfg *StreamConfig) (*Stream, error) {
	err := portaudio.Initialize()
	if err != nil {
		return nil, err
	}

	buffer := make([]int32, cfg.FramesPerBuffer)
	stream, err := portaudio.OpenDefaultStream(
		cfg.InputChannels,
		0,
		cfg.SampleRate,
		cfg.FramesPerBuffer,
		buffer,
	)
	if err != nil {
		return nil, err
	}

	return &Stream{
		cfg:    cfg,
		stream: stream,
		mutex:  &sync.Mutex{},
		state:  Opened,
		buffer: buffer,
	}, nil
}

func (s *Stream) Start() error {
	switch s.state {
	case Opened:
		return s.stream.Start()
	case Closed:
		buffer := make([]int32, s.cfg.FramesPerBuffer)
		stream, err := portaudio.OpenDefaultStream(
			s.cfg.InputChannels,
			0,
			s.cfg.SampleRate,
			s.cfg.FramesPerBuffer,
			buffer,
		)
		if err != nil {
			log.Printf("open default stream error -- %s", err)
			return err
		}

		s.buffer = buffer
		s.stream = stream
		s.state = Started
		stream.Start()

		return nil
	case Started:
		return ErrAlreadyStarted
	}

	return nil
}

func (s *Stream) Close() error {
	err := s.stream.Close()
	if err != nil {
		return err
	}

	s.state = Closed
	return nil
}

func (s *Stream) Read() ([]int32, error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	err := s.stream.Read()
	if err != nil {
		return nil, err
	}

	toRet := make([]int32, s.cfg.FramesPerBuffer)
	copy(toRet, s.buffer)
	return toRet, nil
}

// Terminates portaudio
func (s *Stream) Terminate() {
	portaudio.Terminate()
}
