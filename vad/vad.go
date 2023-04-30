package vad

import (
	"math"
	"sync"

	"github.com/gordonklaus/portaudio"
)

const (
	DefaultSampleInterval   = 500
	DefaultVoiceTimeframe   = 10
	DefaultSilenceTimeframe = 2000
	DefaultInputChannels    = 1
	DefaultSampleRate       = 22050
	DefaultFramesPerBuffer  = 64

	DefaultAMBAVG = 400000000000000.0
)

type VADConfig struct {
	// milliseconds
	VoiceTimeframe   int
	SilenceTimeframe int

	SampleRate      float64
	InputChannels   int
	FramesPerBuffer int
}

// the average has weight towards recent ambient noise
type VAD struct {
	cfg       *VADConfig
	running   bool
	quitCh    chan bool
	pauseCh   chan bool
	resumeCh  chan bool
	active    bool // listening for silence or
	ambAvg    float32
	buffer    []int32
	listeners []VADListener
}

func DefaultVADConfig() *VADConfig {
	return &VADConfig{
		VoiceTimeframe:   DefaultVoiceTimeframe,
		SilenceTimeframe: DefaultSilenceTimeframe,
		SampleRate:       DefaultSampleRate,
		InputChannels:    DefaultInputChannels,
		FramesPerBuffer:  DefaultFramesPerBuffer,
	}
}

func NewVAD(cfg *VADConfig) *VAD {
	return &VAD{
		cfg:       cfg,
		running:   false,
		quitCh:    make(chan bool),
		pauseCh:   make(chan bool),
		resumeCh:  make(chan bool),
		active:    false,
		ambAvg:    DefaultAMBAVG,
		buffer:    make([]int32, cfg.FramesPerBuffer),
		listeners: make([]VADListener, 0),
	}
}

func (v *VAD) RegisterListener(listener VADListener) {
	v.listeners = append(v.listeners, listener)
}

func (v *VAD) Start() {
	go func() {
		stream, err := v.initAudio()
		if err != nil {
			return
		}
		defer v.closeAudio(stream)

		v.running = true
		for {
			quit := false
			select {
			case <-v.pauseCh:
				<-v.resumeCh
			case <-v.quitCh:
				quit = true
			default:
				v.detect(stream)
			}

			if quit {
				break
			}
		}
	}()
}

func (v *VAD) Pause() {
	if v.running {
		v.running = false
		v.pauseCh <- true
	}
}

func (v *VAD) Resume() {
	if !v.running {
		v.running = true
		v.resumeCh <- true
	}
}

// listens to background to check what the threshold should be
// takes in timeframe/ which is in milliseconds
func (v *VAD) sample(stream portaudio.Stream, timeframe int) error {
	for {
		if v.running && !v.active {
			silenceThreshold := float32(0)
			framesDetected := 0
			windowSamples := int(v.cfg.SampleRate / float64(v.cfg.FramesPerBuffer))
			for framesDetected < windowSamples {
				err := stream.Read()
				if err != nil {
					return err
				}

				for j := range v.buffer {
					power := math.Pow(float64(v.buffer[j]), 2)
					silenceThreshold += float32(power)
					framesDetected++
				}
			}
			silenceThreshold /= float32(windowSamples)
			v.ambAvg = (v.ambAvg + silenceThreshold) / 2
		}
	}
}

// the listening for voice window should be smaller than listening for silence window
// because we want to record right after we hear voice, and a little bit of silence will trigger the recorder to stop
func (v *VAD) detect(stream *portaudio.Stream) error {
	var windowSamples int
	if v.active {
		windowSamples = int((float32(v.cfg.SilenceTimeframe) / 1000.0) * float32(v.cfg.SampleRate))
	} else {
		windowSamples = int((float32(v.cfg.VoiceTimeframe) / 1000.0) * float32(v.cfg.SampleRate))
	}
	avgEnergy := float32(0)
	framesDetected := 0
	for framesDetected < windowSamples {
		if !v.running {
			return nil
		}

		err := stream.Read()
		if err != nil {
			return err
		}

		for j := range v.buffer {
			avgEnergy += float32(math.Pow(float64(v.buffer[j]), 2))
			framesDetected++
		}
	}
	avgEnergy /= float32(windowSamples)

	if avgEnergy > float32(v.ambAvg) && !v.active {
		v.notifyListeners(true)
		v.active = true
	} else if avgEnergy <= float32(v.ambAvg) && v.active {
		v.notifyListeners(false)
		v.active = false
	}

	return nil
}

func (v *VAD) notifyListeners(speech bool) {
	var wg sync.WaitGroup
	for i := range v.listeners {
		wg.Add(1)
		go func(index int, speech bool) {
			if speech {
				v.listeners[index].OnSpeechDetected()
			} else {
				v.listeners[index].OnSilenceDetected()
			}
		}(i, speech)
	}
	wg.Wait()
}

func (v *VAD) initAudio() (*portaudio.Stream, error) {
	err := portaudio.Initialize()
	if err != nil {
		return nil, err
	}

	stream, err := portaudio.OpenDefaultStream(
		v.cfg.InputChannels,
		0,
		v.cfg.SampleRate,
		v.cfg.FramesPerBuffer,
		v.buffer,
	)
	if err != nil {
		return nil, err
	}

	err = stream.Start()
	if err != nil {
		return nil, err
	}

	return stream, nil
}

func (v *VAD) closeAudio(stream *portaudio.Stream) {
	portaudio.Terminate()
	stream.Close()
}
