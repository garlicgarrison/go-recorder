package wavseg

import (
	"bytes"
	"math"

	"github.com/garlicgarrison/go-recorder/codec"
)

const (
	DefaultThreshold     = 0.1
	DefaultMinSilentTime = 300 // milliseconds

	DefaultCutoffSpeechInterval = 80 // milliseconds
)

func WavSeg(wav *bytes.Buffer) []*bytes.Buffer {
	w := &codec.WAVFile{}
	err := w.DecodeWAV(wav)
	if err != nil {
		return nil
	}

	threshold := DefaultThreshold * rms(w.Data)
	minSilenceLength := int((float32(DefaultMinSilentTime) / 1000.0) * float32(w.Header.SampleRate))
	var chunks [][]int32
	var chunkStart int
	var chunkEnd int
	var inChunk bool
	var silenceLength int

	for i := 0; i < len(w.Data); i += 2 {
		amplitude := float64(w.Data[i])

		if amplitude > threshold {
			silenceLength = 0
			if !inChunk {
				chunkStart = i
				inChunk = true
			}
		} else {
			silenceLength++
			if inChunk && silenceLength > minSilenceLength {
				chunkEnd = i
				inChunk = false
				chunks = append(chunks, w.Data[chunkStart:chunkEnd])
			}
		}
	}

	if inChunk {
		chunks = append(chunks, w.Data[chunkStart:])
	}

	toRet := []*bytes.Buffer{}
	for _, chunk := range chunks {
		if len(chunk) < int((float32(DefaultCutoffSpeechInterval)/1000.0)*float32(w.Header.SampleRate)) {
			continue
		}

		f := codec.NewDefaultWAV(chunk)
		f.Header = w.Header

		buf, err := f.EncodeWAV()
		if err != nil {
			return nil
		}

		toRet = append(toRet, buf)
	}
	return toRet
}

func rms(chunk []int32) float64 {
	sum := float64(0)
	for _, c := range chunk {
		sum += math.Pow(float64(c), 2)
	}
	return math.Sqrt(sum / float64(len(chunk)))
}
