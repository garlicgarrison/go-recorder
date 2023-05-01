package codec

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"errors"
	"io"
)

var (
	ErrInvalidWAV = errors.New("invalid wav file")
)

type WAVHeader struct {
	RIFF          [4]byte // "RIFF"
	TotalSize     uint32  // Total file size DEFAULT: 36 + 4*numSamples
	Format        [4]byte // "WAVE"
	FMT           [4]byte // "fmt "
	ChunkSize     uint32  // Should be 16 for PCM format DEFAULT: 16
	AudioFormat   uint16  // Should be 1 for PCM format DEFAULT: 1
	NumChannels   uint16  // Mono = 1, Stereo = 2, etc. DEFAULT: 1
	SampleRate    uint32  // Number of samples per second DEFAULT: 44100
	ByteRate      uint32  // Number of bytes per second DEFAULT: 44100 * 4
	BlockAlign    uint16  // Number of bytes per sample DEFAULT: numChannels * 4
	BitsPerSample uint16  // Number of bits per sample DEFAULT: 32
}

type WAVFile struct {
	Header     WAVHeader
	DataHeader [4]byte // DEFAULT: "data"
	DataBytes  uint32  // DEFAULT: 4 * len(stream)
	Data       []int32
}

func NewDefaultWAV(stream []int32) *WAVFile {
	return &WAVFile{
		Header: WAVHeader{
			RIFF:          [4]byte{'R', 'I', 'F', 'F'},
			TotalSize:     uint32(36 + 4*len(stream)),
			Format:        [4]byte{'W', 'A', 'V', 'E'},
			FMT:           [4]byte{'f', 'm', 't', ' '},
			ChunkSize:     16,
			AudioFormat:   1,
			NumChannels:   1,
			SampleRate:    22050,
			ByteRate:      22050 * 4,
			BlockAlign:    4,
			BitsPerSample: 32,
		},
		DataHeader: [4]byte{'d', 'a', 't', 'a'},
		DataBytes:  uint32(4 * len(stream)),
		Data:       stream,
	}
}

func (f *WAVFile) EncodeWAV() (*bytes.Buffer, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	// format
	err := binary.Write(w, binary.LittleEndian, f.Header.RIFF)
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, f.Header.TotalSize) //total size
	if err != nil {
		return nil, err
	}

	// wave declaration
	err = binary.Write(w, binary.LittleEndian, f.Header.Format)
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, f.Header.FMT)
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, f.Header.ChunkSize) //chunk size
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, f.Header.AudioFormat) //format tag
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, f.Header.NumChannels)
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, f.Header.SampleRate)
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, f.Header.ByteRate)
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, f.Header.BlockAlign) //bytes per sample
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, f.Header.BitsPerSample) //bits per sample
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, f.DataHeader)
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.LittleEndian, int32(f.Header.TotalSize-36))
	if err != nil {
		return nil, err
	}

	err = writeRawAudio(w, false, f.Data)
	if err != nil {
		return nil, err
	}

	err = w.Flush()
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

func (f *WAVFile) DecodeWAV(buf *bytes.Buffer) error {
	// read WAV header fields
	err := binary.Read(buf, binary.LittleEndian, &f.Header)
	if err != nil {
		return err
	}

	// check if file format is supported
	if string(f.Header.RIFF[:]) != "RIFF" || string(f.Header.Format[:]) != "WAVE" || string(f.Header.FMT[:]) != "fmt " || f.Header.AudioFormat != 1 {
		return ErrInvalidWAV
	}

	// skips any extra bytes in the format subchunk
	if f.Header.ChunkSize > 16 {
		_, err = io.CopyN(io.Discard, buf, int64(f.Header.ChunkSize-16))
		if err != nil {
			return err
		}
	}

	var dataHeader [4]byte
	err = binary.Read(buf, binary.LittleEndian, &dataHeader)
	if err != nil {
		return err
	}
	f.DataHeader = dataHeader

	var dataSize uint32
	err = binary.Read(buf, binary.LittleEndian, &dataSize)
	if err != nil {
		return err
	}
	f.DataBytes = dataSize

	data := make([]int32, (f.Header.TotalSize-36)/4)
	err = binary.Read(buf, binary.LittleEndian, &data)
	if err != nil {
		return err
	}
	f.Data = data

	return nil
}

type AIFFHeader struct {
	FORM           [4]byte // "FORM"
	TotalSize      uint32  // Total file size DEFAULT: 4 + 8 + 18 + 8 + 8 + 4*numSamples
	FormType       [4]byte // "AIFF"
	AudioFormat    [4]byte // "COMM"
	ChunkSize      uint32  // DEFAULT: 18
	NumChannels    uint16  // Mono = 1, Stereo = 2, etc. DEFAULT: 1
	NumSamples     uint32  // number of sample frames
	BitsPerSample  uint16  // Number of bits per sample DEFAULT: 32
	SampleRate     float32 // Number of samples per second DEFAULT: 44100.0
	SSND           [4]byte // DEFAULT: "SSND"
	SoundChunkSize uint32  // DEFAULT: 4*numSamples+8
	Offset         uint32  // DEFAULT: 0
	Block          uint32  // DEFAULT: 0
}

type AIFFFile struct {
	Header AIFFHeader
	Data   []int32
}

func NewDefaultAIFF(stream []int32) *AIFFFile {
	return &AIFFFile{
		Header: AIFFHeader{
			FORM:           [4]byte{'F', 'O', 'R', 'M'},
			TotalSize:      uint32(4 + 8 + 18 + 8 + 8 + 4*len(stream)),
			FormType:       [4]byte{'A', 'I', 'F', 'F'},
			AudioFormat:    [4]byte{'C', 'O', 'M', 'M'},
			ChunkSize:      18,
			NumChannels:    1,
			NumSamples:     uint32(len(stream)),
			BitsPerSample:  32,
			SampleRate:     22050.0,
			SSND:           [4]byte{'S', 'S', 'N', 'D'},
			SoundChunkSize: uint32(4*len(stream) + 8),
			Offset:         0,
			Block:          0,
		},
		Data: stream,
	}
}

func (f *AIFFFile) EncodeAIFF() (*bytes.Buffer, error) {
	var buf bytes.Buffer
	w := bufio.NewWriter(&buf)

	// format
	err := binary.Write(w, binary.BigEndian, f.Header.FORM)
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.BigEndian, f.Header.TotalSize)
	if err != nil {
		return nil, err
	}

	// aiff declaration
	err = binary.Write(w, binary.BigEndian, f.Header.FormType)
	if err != nil {
		return nil, err
	}

	// comm
	err = binary.Write(w, binary.BigEndian, f.Header.AudioFormat)
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.BigEndian, f.Header.ChunkSize) //size
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.BigEndian, f.Header.NumChannels) //channels
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.BigEndian, f.Header.NumSamples) //samples
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.BigEndian, f.Header.BitsPerSample) //bits per sample
	if err != nil {
		return nil, err
	}

	_, err = w.Write([]byte{0x40, 0x0e, byte(int(f.Header.SampleRate) >> 8), byte(int(f.Header.SampleRate) & 0xFF), 0, 0, 0, 0, 0, 0}) //80-bits 44100 sample rate
	if err != nil {
		return nil, err
	}

	// sound chunk
	err = binary.Write(w, binary.BigEndian, f.Header.SSND)
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.BigEndian, f.Header.SoundChunkSize) //size
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.BigEndian, f.Header.Offset) //offset
	if err != nil {
		return nil, err
	}

	err = binary.Write(w, binary.BigEndian, f.Header.Block) //block
	if err != nil {
		return nil, err
	}

	err = writeRawAudio(w, true, f.Data)
	if err != nil {
		return nil, err
	}

	err = w.Flush()
	if err != nil {
		return nil, err
	}

	return &buf, nil
}

func writeRawAudio(w *bufio.Writer, endian bool, fullStream []int32) error {
	for _, frame := range fullStream {
		var err error
		if endian {
			err = binary.Write(w, binary.BigEndian, frame)
		} else {
			err = binary.Write(w, binary.LittleEndian, frame)
		}
		if err != nil {
			return err
		}
	}

	return nil
}
