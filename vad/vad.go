package vad

import (
	"log"
	"math"
)

type Detection string

const (
	DefaultSampleInterval   = 500
	DefaultVoiceTimeframe   = 5
	DefaultSilenceTimeframe = 12
	DefaultInputChannels    = 1
	DefaultSampleRate       = 22050
	DefaultFramesPerBuffer  = 64

	DefaultAMBAVG                  = 500000000.0
	DefaultSpeechAMBAVGMultiplier  = 2
	DefaultSilenceAMBAVGMultiplier = 1 // max: 1

	NewSilenceThresholdWeight = 0.3 // max: 1
	AudioChanSize             = 64

	SpeechDetection  Detection = "speech"
	SilenceDetection Detection = "silence"
)

type VADConfig struct {
	// milliseconds
	VoiceTimeframe    int
	SilenceTimeframe  int
	SamplingTimeframe int

	SampleRate      float64
	InputChannels   int
	FramesPerBuffer int
}

// the average has weight towards recent ambient noise
type VAD struct {
	cfg            *VADConfig
	ambAvg         float32
	audioChan      chan []int32
	speechWindows  int
	silenceWindows int
	sampleWindows  int
}

func DefaultVADConfig() *VADConfig {
	return &VADConfig{
		VoiceTimeframe:    DefaultVoiceTimeframe,
		SilenceTimeframe:  DefaultSilenceTimeframe,
		SamplingTimeframe: DefaultSampleInterval,
		SampleRate:        DefaultSampleRate,
		InputChannels:     DefaultInputChannels,
		FramesPerBuffer:   DefaultFramesPerBuffer,
	}
}

func NewVAD(cfg *VADConfig) *VAD {
	return &VAD{
		cfg:            cfg,
		ambAvg:         DefaultAMBAVG,
		audioChan:      make(chan []int32, AudioChanSize),
		speechWindows:  int((float32(cfg.VoiceTimeframe) / 1000.0) * float32(cfg.SampleRate)),
		silenceWindows: int((float32(cfg.SilenceTimeframe) / 1000.0) * float32(cfg.SampleRate)),
		sampleWindows:  int((float32(cfg.SamplingTimeframe) / 1000.0) * float32(cfg.SampleRate)),
	}
}

func (v *VAD) AddBuffer(b []int32) {
	select {
	case v.audioChan <- b:
	default:
		log.Printf("channel is full -- discarding value")
	}
}

func (v *VAD) DetectSpeech(stop chan bool) bool {
	buffers := [][]int32{}
	for len(buffers) < v.speechWindows {
		select {
		case newBuf := <-v.audioChan:
			buffers = append(buffers, newBuf)
		case <-stop:
			return false
		default:
			continue
		}
	}

	isSpeech := v.detect(SpeechDetection, buffers)
	if !isSpeech {
		v.sample(buffers)
	}

	return isSpeech
}

func (v *VAD) DetectSilence(stop chan bool) bool {
	buffers := [][]int32{}
	for len(buffers) < v.silenceWindows {
		select {
		case newBuf := <-v.audioChan:
			buffers = append(buffers, newBuf)
		case <-stop:
			return false
		default:
			continue
		}
	}

	return v.detect(SilenceDetection, buffers)
}

// the listening for voice window should be smaller than listening for silence window
// because we want to record right after we hear voice, and a little bit of silence will trigger the recorder to stop
// this function should have the only stream reads
func (v *VAD) detect(detection Detection, buffers [][]int32) bool {
	avgEnergy := float32(0)
	for _, buffer := range buffers {
		for _, amp := range buffer {
			avgEnergy += float32(math.Pow(float64(amp), 2))
		}
	}
	avgEnergy = float32(math.Sqrt(float64(avgEnergy / float32(len(buffers)))))

	if avgEnergy > float32(v.ambAvg)*DefaultSpeechAMBAVGMultiplier &&
		detection == SpeechDetection {
		return true
	} else if avgEnergy <= float32(v.ambAvg)*DefaultSilenceAMBAVGMultiplier &&
		detection == SilenceDetection {
		return true
	}

	return false
}

func (v *VAD) sample(buffers [][]int32) {
	silenceThreshold := float32(0)
	for _, buffer := range buffers {
		for j := range buffer {
			silenceThreshold += float32(math.Pow(float64(buffer[j]), 2))
		}
	}
	silenceThreshold = float32(math.Sqrt(float64(silenceThreshold / float32(len(buffers)))))
	v.ambAvg = v.ambAvg*(1-NewSilenceThresholdWeight) + silenceThreshold*NewSilenceThresholdWeight
}
