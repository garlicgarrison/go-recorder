package main

import (
	"bytes"
	"log"
	"os"

	"github.com/garlicgarrison/go-recorder/wavseg"
)

func main() {
	wavBytes, err := os.ReadFile("wav_test.wav")
	if err != nil {
		// log.Printf("read error -- %s", err)
		return
	}

	buffers := wavseg.WavSeg(bytes.NewBuffer(wavBytes))
	log.Printf("buffers %d", len(buffers))
}
