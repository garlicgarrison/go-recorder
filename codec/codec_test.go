package codec

import (
	"bytes"
	"log"
	"math"
	"os"
	"testing"
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

func TestEncodeDecodeWAV(t *testing.T) {
	waves := getTestData(2)
	waves = append(waves, getTestDataSilence(1)...)
	waves = append(waves, getTestData(2)...)

	wav := NewDefaultWAV(waves)
	b, err := wav.EncodeWAV()
	if err != nil {
		t.Fatalf("encoding error - %s", err)
	}

	f, err := os.Create("tests/wav_test.wav")
	if err != nil {
		t.Fatalf("file creation error - %s", err)
	}
	defer f.Close()

	_, err = f.Write(b.Bytes())
	if err != nil {
		t.Fatalf("write error - %s", err)
	}

	wavBytes, err := os.ReadFile("tests/wav_test.wav")
	if err != nil {
		t.Fatalf("read error -- %s", err)
	}

	log.Printf("len %d", len(wavBytes))
	err = wav.DecodeWAV(bytes.NewBuffer(wavBytes))
	if err != nil {
		t.Fatalf("decode error -- %s", err)
	}
}
