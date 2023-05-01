package stream

import (
	"sync"

	"github.com/gordonklaus/portaudio"
)

const (
	DefaultInputChannels   = 1
	DefaultSampleRate      = 22050
	DefaultFramesPerBuffer = 64
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
	quit   chan bool
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

	pbuffer := make([]int32, cfg.FramesPerBuffer)
	rbuffer := make([]int32, cfg.FramesPerBuffer)

	stream, err := portaudio.OpenDefaultStream(
		cfg.InputChannels,
		0,
		cfg.SampleRate,
		cfg.FramesPerBuffer,
		pbuffer,
		rbuffer,
	)
	if err != nil {
		return nil, err
	}

	err = stream.Start()
	if err != nil {
		return nil, err
	}

	return &Stream{
		cfg:    cfg,
		stream: stream,
		mutex:  &sync.Mutex{},
		buffer: pbuffer,
	}, nil
}

func (s *Stream) Close() error {
	s.quit <- true
	portaudio.Terminate()
	return s.stream.Close()
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
