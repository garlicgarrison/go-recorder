package recorder

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"time"

	"github.com/gordonklaus/portaudio"
)

type Format string

const (
	WAV  Format = "wav"
	AIFF Format = "aiff"
)

var (
	ErrInvalidRecorderConfig = errors.New("invalid config")
)

type RecorderConfig struct {
	SampleRate      float64
	InputChannels   int
	FramesPerBuffer int
	MaxTime         int //milliseconds
}

type Recorder struct {
	cfg    *RecorderConfig
	buffer []int32
}

func NewRecorder(cfg *RecorderConfig) (*Recorder, error) {
	if cfg == nil {
		return nil, ErrInvalidRecorderConfig
	}

	return &Recorder{
		cfg:    cfg,
		buffer: make([]int32, cfg.FramesPerBuffer),
	}, nil
}

func (r *Recorder) Record(format Format, quit chan os.Signal) (*bytes.Buffer, error) {
	timerChan := make(chan bool)
	go func() {
		time.Sleep(time.Millisecond * time.Duration(r.cfg.MaxTime))
		timerChan <- true
	}()

	err := portaudio.Initialize()
	if err != nil {
		return nil, err
	}

	defer portaudio.Terminate()

	stream, err := portaudio.OpenDefaultStream(
		r.cfg.InputChannels,
		0,
		r.cfg.SampleRate,
		r.cfg.FramesPerBuffer,
		r.buffer,
	)
	if err != nil {
		return nil, err
	}

	err = stream.Start()
	if err != nil {
		return nil, err
	}

	defer stream.Close()

	fullStream := [][]int32{}
	for {
		err = stream.Read()
		if err != nil {
			return nil, err
		}

		currStream := make([]int32, r.cfg.FramesPerBuffer)
		copy(currStream, r.buffer)
		fullStream = append(fullStream, currStream)
		select {
		case <-quit:
			if err = stream.Stop(); err != nil {
				return nil, err
			}

			switch format {
			case AIFF:
				return r.writeAIFF(fullStream)
			case WAV:
				return r.writeWAV(fullStream)
			}
		case <-timerChan:
			if err = stream.Stop(); err != nil {
				return nil, err
			}

			switch format {
			case AIFF:
				return r.writeAIFF(fullStream)
			case WAV:
				return r.writeWAV(fullStream)
			}
		default:
		}
	}
}

func (r *Recorder) writeWAV(fullStream [][]int32) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	// format
	err := binary.Write(w, binary.LittleEndian, []byte("RIFF"))
	if err != nil {
		return nil, err
	}

	numSamples := r.cfg.FramesPerBuffer * len(fullStream)
	totalBytes := 36 + 4*numSamples
	err = binary.Write(w, binary.LittleEndian, int32(totalBytes)) //total size
	if err != nil {
		return nil, err
	}

	// wave declaration
	err = binary.Write(w, binary.LittleEndian, []byte("WAVEfmt "))
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, int32(16)) //chunk size
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, int16(1)) //format tag
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, int16(r.cfg.InputChannels))
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, int32(r.cfg.SampleRate))
	if err != nil {
		return nil, err
	}

	bytesPerSec := 4 * r.cfg.SampleRate
	err = binary.Write(w, binary.LittleEndian, int32(bytesPerSec))
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, int16(r.cfg.InputChannels*4)) //bytes per sample
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, int16(32)) //bits per sample
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, []byte("data"))
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, int32(totalBytes-36))
	if err != nil {
		return nil, err
	}

	err = r.writeRawAudio(w, false, fullStream)
	if err != nil {
		return nil, err
	}

	err = w.Flush()
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

func (r *Recorder) writeAIFF(fullStream [][]int32) (*bytes.Buffer, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	// format
	_, err := w.WriteString("FORM")
	if err != nil {
		return nil, err
	}

	numSamples := r.cfg.FramesPerBuffer * len(fullStream)
	totalBytes := 4 + 8 + 18 + 8 + 8 + 4*numSamples
	err = binary.Write(w, binary.BigEndian, int32(totalBytes)) // total size
	if err != nil {
		return nil, err
	}

	// aiff declaration
	_, err = w.WriteString("AIFF")
	if err != nil {
		return nil, err
	}

	// comm
	_, err = w.WriteString("COMM")
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.BigEndian, int32(18)) //size
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.BigEndian, int16(1)) //channels
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.BigEndian, int32(numSamples)) //samples
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.BigEndian, int16(32)) //bits per sample
	if err != nil {
		return nil, err
	}

	_, err = w.Write([]byte{0x40, 0x0e, 0xac, 0x44, 0, 0, 0, 0, 0, 0}) //80-bits 44100 sample rate
	if err != nil {
		return nil, err
	}

	// sound chunk
	_, err = w.WriteString("SSND")
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.BigEndian, int32(4*numSamples+8)) //size
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.BigEndian, int32(0)) //offset
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.BigEndian, int32(0)) //block
	if err != nil {
		return nil, err
	}

	err = r.writeRawAudio(w, true, fullStream)
	if err != nil {
		return nil, err
	}

	err = w.Flush()
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

func (r *Recorder) writeRawAudio(w *bufio.Writer, endian bool, fullStream [][]int32) error {
	for _, chunk := range fullStream {
		var err error
		if endian {
			err = binary.Write(w, binary.BigEndian, chunk)
		} else {
			err = binary.Write(w, binary.LittleEndian, chunk)
		}
		if err != nil {
			return err
		}
	}

	return nil
}
