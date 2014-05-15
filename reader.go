package wav

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"time"
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
	duration       time.Duration

	samplesRead uint32
	numSamples  uint32
}

func (wav WavReader) String() string {
	msg := fmt.Sprintln("File informations")
	msg += fmt.Sprintln("=================")
	msg += fmt.Sprintf("File size         : %d bytes\n", wav.size)
	msg += fmt.Sprintf("Canonical format  : %v\n", wav.canonical && !wav.extraChunk)
	// chunk fmt
	msg += fmt.Sprintf("Audio format      : %d\n", wav.chunkFmt.AudioFormat)
	msg += fmt.Sprintf("Number of channels: %d\n", wav.chunkFmt.NumChannels)
	msg += fmt.Sprintf("Sampling rate     : %d Hz\n", wav.chunkFmt.SampleRate)
	msg += fmt.Sprintf("Sample size       : %d bits\n", wav.chunkFmt.BitsPerSample)
	// calculated
	msg += fmt.Sprintf("Number of samples : %d\n", wav.numSamples)
	msg += fmt.Sprintf("Sound size        : %d bytes\n", wav.dataBlocSize)
	msg += fmt.Sprintf("Sound duration    : %v\n", wav.duration)

	return msg
}

func NewWavReader(rd io.ReadSeeker, size int64) (wav *WavReader, err error) {
	if size > maxSize {
		return nil, ErrInputToLarge
	}

	wav = new(WavReader)
	wav.input = rd
	wav.size = size

	err = wav.parseHeaders()
	if err != nil {
		return nil, err
	}

	wav.samplesRead = 0

	return wav, nil
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

	if wav.header.Ftype != tokenRiff {
		return ErrNotRiff
	}

	if wav.header.ChunkSize+8 != uint32(wav.size) {
		return ErrIncorrectChunkSize{wav.header.ChunkSize + 8, uint32(wav.size)}
	}

	if wav.header.ChunkFormat != tokenWaveFormat {
		return ErrNotWave
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

		switch chunk {
		case tokenChunkFmt:
			// seek 4 bytes back because riffChunkFmt reads the chunkSize again
			if _, err = wav.input.Seek(-4, os.SEEK_CUR); err != nil {
				return err
			}
			wav.canonical = chunkSize == 16 // canonical format if chunklen == 16
			if err = wav.parseChunkFmt(); err != nil {
				return err
			}
		case tokenData:
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
		return ErrFormatNotSupported
	}

	wav.bytesPerSample = uint32(wav.chunkFmt.BitsPerSample / 8)
	wav.numSamples = wav.dataBlocSize / wav.bytesPerSample
	wav.duration = time.Duration(float64(wav.numSamples)/float64(wav.chunkFmt.SampleRate)) * time.Second

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

func (wav *WavReader) GetSampleCount() uint32 {
	return wav.numSamples
}

func (w WavReader) GetWavFile() WavFile {
	return WavFile{
		SampleRate:      w.chunkFmt.SampleRate,
		Channels:        w.chunkFmt.NumChannels,
		SignificantBits: w.chunkFmt.BitsPerSample,
	}
}

func (wav *WavReader) ReadRawSample() ([]byte, error) {
	if wav.samplesRead > wav.numSamples {
		return nil, io.EOF
	}

	buf := make([]byte, wav.bytesPerSample)
	n, err := wav.input.Read(buf)
	if err != nil {
		return nil, err
	}

	if n != int(wav.bytesPerSample) {
		return nil, fmt.Errorf("Read %d bytes, should have read %d", n, wav.bytesPerSample)
	}

	wav.samplesRead += 1
	// w.bytesWritten += n

	return buf, nil
}

func (wav *WavReader) ReadSample() (n int32, err error) {
	s, err := wav.ReadRawSample()
	if err != nil {
		return 0, err
	}

	switch wav.bytesPerSample {
	case 1:
		n = int32(s[0])
	case 2:
		n = int32(s[0]) + int32(s[1])<<8
	case 3:
		n = int32(s[0]) + int32(s[1])<<8 + int32(s[2])<<16
	case 4:
		n = int32(s[0]) + int32(s[1])<<8 + int32(s[2])<<16 + int32(s[3])<<24
	default:
		n = 0
		err = fmt.Errorf("Unhandled bytesPerSample! b:%d", wav.bytesPerSample)
	}

	return
}
