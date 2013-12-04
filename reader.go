package wav

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
)

const (
	maxWavSize = 2 << 31
)

type WavReader struct {
	input io.ReadSeeker
	size  int64

	header   *riffHeader
	chunkFmt *riffChunkFmt

	canonical      bool
	extraChunk     bool
	firstSamplePos uint32
	dataBlocSize   uint32
	bytesPerSample uint32
	numSamples     uint32
	duration       time.Duration
}

func (wav WavReader) String() string {
	msg := fmt.Sprintln("File informations")
	msg += fmt.Sprintln("=================")
	msg += fmt.Sprintf("File size         : %d bytes\n", wav.size)
	msg += fmt.Sprintf("Canonical format  : %v\n", wav.canonical && !wav.extraChunk)
	msg += fmt.Sprintf("Audio format      : %d\n", wav.chunkFmt.AudioFormat)
	msg += fmt.Sprintf("Number of channels: %d\n", wav.chunkFmt.NumChannels)
	msg += fmt.Sprintf("Sampling rate     : %d Hz\n", wav.chunkFmt.SampleFreq)
	msg += fmt.Sprintf("Sample size       : %d bits\n", wav.chunkFmt.BitsPerSample)
	msg += fmt.Sprintf("Number of samples : %d\n", wav.numSamples)
	msg += fmt.Sprintf("Sound size        : %d bytes\n", wav.dataBlocSize)
	msg += fmt.Sprintf("Sound duration    : %v\n", wav.duration)

	return msg
}

type riffHeader struct {
	Ftype       [4]byte
	ChunkSize   uint32
	ChunkFormat [4]byte
}

type riffChunkFmt struct {
	AudioFormat   uint16 // 1 = PCM not compressed
	NumChannels   uint16
	SampleFreq    uint32
	BytesPerSec   uint32
	BytesPerBloc  uint16
	BitsPerSample uint16
}

func NewWavReader(rd io.ReadSeeker, size int64) (wav *WavReader, err error) {
	if size > maxWavSize {
		err = fmt.Errorf("Input too large")
		return
	}

	wav = &WavReader{}
	wav.input = rd
	wav.size = size

	err = wav.parseHeaders()

	return
}

func (wav *WavReader) parseHeaders() (err error) {

	wav.header = &riffHeader{}
	var (
		chunk     [4]byte
		chunkSize uint32
	)

	// decode header
	if err = binary.Read(wav.input, binary.LittleEndian, wav.header); err != nil {
		return err
	}

	if string(wav.header.Ftype[:]) != "RIFF" {
		return fmt.Errorf("Not a RIFF file")
	}

	if wav.header.ChunkSize+8 != uint32(wav.size) {
		return fmt.Errorf("Damaged file. Chunk size != file size.")
	}

	if string(wav.header.ChunkFormat[:]) != "WAVE" {
		return fmt.Errorf("Not a WAVE file")
	}

readLoop:
	for {
		// Read next chunkID
		if err = binary.Read(wav.input, binary.BigEndian, &chunk); err != nil {
			return err
		}
		// and it's size in bytes
		if err = binary.Read(wav.input, binary.LittleEndian, &chunkSize); err != nil {
			return err
		}

		switch string(chunk[:4]) {
		case "fmt ":
			wav.canonical = chunkSize == 16 // canonical format if chunklen == 16
			if err = wav.parseChunkFmt(); err != nil {
				return err
			}
		case "data":
			size, _ := wav.input.Seek(0, os.SEEK_CUR)
			wav.firstSamplePos = uint32(size)
			wav.dataBlocSize = uint32(chunkSize)
			break readLoop
		default:
			//fmt.Fprintf(os.Stderr, "Skip unused chunk \"%s\" (%d bytes).\n", chunk, chunkSize)
			wav.extraChunk = true
			if _, err = wav.input.Seek(int64(chunkSize), os.SEEK_CUR); err != nil {
				return err
			}
		}
	}

	// Is audio supported ?
	if wav.chunkFmt.AudioFormat != 1 {
		return fmt.Errorf("Only PCM (not compressed) format is supported.")
	}

	wav.numSamples = wav.dataBlocSize / uint32(wav.chunkFmt.BitsPerSample>>3)
	wav.duration = time.Duration(float64(wav.dataBlocSize) / float64(wav.chunkFmt.BitsPerSample>>3))

	return nil
}

// parseChunkFmt
func (wav *WavReader) parseChunkFmt() (err error) {
	wav.chunkFmt = &riffChunkFmt{}

	if err = binary.Read(wav.input, binary.LittleEndian, wav.chunkFmt); err != nil {
		return err
	}

	if wav.canonical == false {
		var extraparams uint32
		// Get extra params size
		if err = binary.Read(wav.input, binary.LittleEndian, &extraparams); err != nil {
			return err
		}
		// Skip them
		if _, err = wav.input.Seek(int64(extraparams), os.SEEK_CUR); err != nil {
			return err
		}
	}

	return nil
}
