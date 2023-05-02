package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/garlicgarrison/go-recorder/recorder"
	"github.com/garlicgarrison/go-recorder/stream"
	"github.com/garlicgarrison/go-recorder/vad"
)

type Voice struct {
	kill chan bool
	r    *recorder.Recorder
	vad  *vad.VAD
}

// func NewVoice(r *recorder.Recorder) *Voice {
// 	// sig := make(chan os.Signal, 1)
// 	// signal.Notify(sig, os.Interrupt, os.Kill)

// 	return &Voice{
// 		openAI: client,
// 		kill:   make(chan bool),
// 		r:      r,
// 	}
// }

// func (v *Voice) Start() {

// 	for {
// 		buffer, err := v.r.RecordVAD(recorder.WAV)
// 		if err != nil {
// 			log.Fatalf("recording error -- %s", err)
// 		}

// 		segments := wavseg.WavSeg(buffer)
// 		if segments == nil {
// 			continue
// 		}
// 		log.Printf("num of segments: %d", len(segments))

// 		// ctx := context.Background()
// 		transcriptions := make([]string, len(segments))
// 		// errChan := make(chan error, len(segments))
// 		//var wg sync.WaitGroup
// 		for id, b := range segments {
// 			log.Printf("%d", id)
// 			//wg.Add(1)
// 			file, err := os.CreateTemp("tmp", "transcribe_")
// 			if err != nil {
// 				log.Fatalf("temp file error -- %s", err)
// 			}

// 			_, err = file.Write(b.Bytes())
// 			if err != nil {
// 				log.Fatalf("file write error -- %s", err)
// 			}

// 			// go func(id int) {
// 			// 	defer wg.Done()

// 			// 	res, err := v.openAI.CreateTranscription(ctx, openai.AudioRequest{
// 			// 		Model:       openai.Whisper1,
// 			// 		FilePath:    file.Name(),
// 			// 		Temperature: 0.5,
// 			// 	})
// 			// 	if err != nil {
// 			// 		errChan <- err
// 			// 		return
// 			// 	}

// 			// 	transcriptions[id] = res.Text
// 			// }(id)
// 		}
// 		//wg.Wait()

// 		prompt := strings.Join(transcriptions, "")
// 		log.Printf("prompt -- %s", prompt)

// 		quit := false
// 		select {
// 		case <-v.kill:
// 			quit = true
// 		default:
// 			continue
// 		}
// 		if quit {
// 			break
// 		}
// 	}

// }

// func (v *Voice) Close() {
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
		default:
			continue
		}
		if quit {
			break
		}
	}

}
