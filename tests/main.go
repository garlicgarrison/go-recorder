package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/garlicgarrison/go-recorder/recorder"
	"github.com/garlicgarrison/go-recorder/stream"
)

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
		}
		if quit {
			break
		}
	}

}
