package codec

import (
	"bytes"
	"log"
	"math"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

func TestEncodeDecodeWAV(t *testing.T) {
	waves := getTestData(2)

	wav := NewDefaultWAV(waves)
	c := &Codec{}
	b, err := c.EncodeWAV(wav)
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
	decodedWAV, err := c.DecodeWAV(bytes.NewBuffer(wavBytes))
	if err != nil {
		t.Fatalf("decode error -- %s", err)
	}

	assert.Equal(t, decodedWAV, wav)
}
