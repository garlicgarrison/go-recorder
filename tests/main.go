package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/garlicgarrison/go-recorder/recorder"
	"github.com/garlicgarrison/go-recorder/stream"
)

// type Voice struct {
// 	kill chan bool
// 	r    *recorder.Recorder
// 	vad  *vad.VAD
// }

// func NewVoice(r *recorder.Recorder, vad *vad.VAD) *Voice {
// 	// sig := make(chan os.Signal, 1)
// 	// signal.Notify(sig, os.Interrupt, os.Kill)

// 	return &Voice{
// 		kill: make(chan bool),
// 		r:    r,
// 		vad:  vad,
// 	}
// }

// func (v *Voice) Start() {
// 	v.vad.RegisterListener(v)
// 	v.vad.Start()
// }

// func (v *Voice) Stop() {
// 	v.vad.Stop()
// }

// func (v *Voice) OnSpeechDetected() {
// 	log.Printf("hi")
// 	buf, err := v.r.Record(recorder.WAV, v.kill)
// 	if err != nil {
// 		log.Printf("recording error -- %s", err)
// 		return
// 	}

// 	log.Printf("done")

// 	buffers := wavseg.WavSeg(buf)
// 	if buffers == nil {
// 		return
// 	}

// 	// transcriptions := make([]string, len(buffers))
// 	var wg sync.WaitGroup
// 	for id, b := range buffers {
// 		wg.Add(1)

// 		go func(index int, buf *bytes.Buffer) {
// 			defer wg.Done()
// 			file, err := os.Create(fmt.Sprintf("transcribe_%d.wav", index))
// 			if err != nil {
// 				return
// 			}

// 			_, err = file.Write(buf.Bytes())
// 			if err != nil {
// 				return
// 			}
// 		}(id, b)
// 	}
// 	wg.Wait()
// }

// func (v *Voice) OnSilenceDetected() {
// 	log.Printf("silence")
// 	v.kill <- true
// }

func main() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, os.Kill)

	stream, err := stream.NewStream(stream.DefaultStreamConfig())
	if err != nil {
		log.Fatalf("stream error -- %s", err)
	}
	defer stream.Close()

	rec, err := recorder.NewRecorder(recorder.DefaultRecorderConfig(), stream)
	if err != nil {
		log.Fatalf("recorder error -- %s", err)
	}

	// vad := vad.NewVAD(vad.DefaultVADConfig(), stream)

	// voice := NewVoice(rec, vad)
	// voice.Start()

	for {
		recording, err := rec.RecordVAD(recorder.WAV)
		if err != nil {
			return
		}

		file, err := os.Create(fmt.Sprintf("transcribe.wav"))
		if err != nil {
			return
		}

		_, err = file.Write(recording.Bytes())
		if err != nil {
			return
		}

		quit := false
		select {
		case <-sig:
			quit = true
			os.Exit(1)
		}
		if quit {
			break
		}
	}

}
