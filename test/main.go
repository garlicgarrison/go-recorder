package main

import (
	"time"

	"github.com/garlicgarrison/go-recorder/vad"
)

const (
	SampleRate = 16000
	FrameSize  = 64
	WindowLen  = 1000 * time.Millisecond
	Threshold  = 0.1
)

func main() {
	// rCfg := recorder.NewDefaultRecorderConfig()
	// _, err := recorder.NewRecorder(rCfg)
	// if err != nil {
	// 	panic(err)
	// }

	// err = portaudio.Initialize()
	// if err != nil {
	// 	panic(err)
	// }

	// defer portaudio.Terminate()

	// buffer := make([]int32, 64)
	// stream, err := portaudio.OpenDefaultStream(
	// 	rCfg.InputChannels,
	// 	0,
	// 	rCfg.SampleRate,
	// 	rCfg.FramesPerBuffer,
	// 	buffer,
	// )
	// if err != nil {
	// 	panic(err)
	// }

	// err = stream.Start()
	// if err != nil {
	// 	panic(err)
	// }

	// defer stream.Close()

	// //startChan := make(chan bool)

	// // get a sample of surrounding sound
	// silenceThreshold := float32(0)
	// i := 0
	// now := time.Now()
	// for {
	// 	err = stream.Read()
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	curr := float32(0)
	// 	for i := range buffer {
	// 		curr += float32(math.Pow(float64(buffer[i]), 2))
	// 	}
	// 	silenceThreshold = (silenceThreshold*float32(i) + float32(curr)) / float32(i+1)
	// 	i++

	// 	if i == 100 {
	// 		break
	// 	}
	// }
	// log.Printf("time %d", time.Since(now)/1000000)

	// for {
	// 	err = stream.Read()
	// 	if err != nil {
	// 		panic(err)
	// 	}

	// 	// Compute average energy over specified window length
	// 	avgEnergy := float32(0)
	// 	windowSamples := int(WindowLen.Seconds() * SampleRate / float64(FrameSize))
	// 	for i := 0; i < windowSamples; i++ {
	// 		avgEnergy += float32(math.Pow(float64(buffer[i]), 2))
	// 	}
	// 	avgEnergy /= float32(windowSamples)
	// 	if avgEnergy > float32(silenceThreshold) {
	// 		log.Printf("avgEnergy %f, silcence %f", avgEnergy, silenceThreshold)
	// 		log.Printf("detected")
	// 	}
	// }

	v := vad.NewVAD(vad.DefaultVADConfig())
	v.Start()
}
