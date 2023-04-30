package wavseg

import (
	"bytes"
	"fmt"
	"math"
	"os"
	"testing"

	"github.com/garlicgarrison/go-recorder/codec"
	"github.com/stretchr/testify/assert"
)

// middle c
func getTestData(seconds int) []int32 {
	amplitude := float64(1<<31 - 1)
	freq := 261.63

	numSamples := 44100 * seconds // 1 second
	testData := make([]int32, numSamples)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(44100)
		testData[i] = int32(amplitude * math.Sin(2*math.Pi*freq*t))
	}

	return testData
}

func getTestDataSilence(seconds int) []int32 {
	amplitude := float64(1<<22 - 1)
	freq := 261.63

	numSamples := 44100 * seconds // 1 second
	testData := make([]int32, numSamples)
	for i := 0; i < numSamples; i++ {
		t := float64(i) / float64(44100)
		testData[i] = int32(amplitude * math.Sin(2*math.Pi*freq*t))
	}

	return testData
}

func TestWavSeg(t *testing.T) {
	waves := getTestData(2)
	waves = append(waves, getTestDataSilence(1)...)
	waves = append(waves, getTestData(2)...)

	wav := codec.NewDefaultWAV(waves)
	b, err := wav.EncodeWAV()
	if err != nil {
		t.Fatalf("encoding error - %s", err)
	}

	buffers := WavSeg(b)
	assert.Equal(t, len(buffers), 2)
	for i := range buffers {
		f, err := os.Create(fmt.Sprintf("tests/wav_test_%d.wav", i))
		if err != nil {
			t.Fatalf("file creation error - %s", err)
		}
		defer f.Close()

		_, err = f.Write(buffers[i].Bytes())
		if err != nil {
			t.Fatalf("write error - %s", err)
		}
	}

	wavBytes, err := os.ReadFile("tests/me.wav")
	if err != nil {
		t.Fatalf("read error -- %s", err)
	}
	buffers = WavSeg(bytes.NewBuffer(wavBytes))
	for i := range buffers {
		f, err := os.Create(fmt.Sprintf("tests/me_test_%d.wav", i))
		if err != nil {
			t.Fatalf("file creation error - %s", err)
		}
		defer f.Close()

		_, err = f.Write(buffers[i].Bytes())
		if err != nil {
			t.Fatalf("write error - %s", err)
		}
	}
}
