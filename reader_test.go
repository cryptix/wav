package wav

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	riff             = []byte{0x52, 0x49, 0x46, 0x46} // "RIFF"
	chunkSize24      = []byte{0x24, 0x00, 0x00, 0x00} // chunkSize
	wave             = []byte{0x57, 0x41, 0x56, 0x45} // "WAVE"
	fmt20            = []byte{0x66, 0x6d, 0x74, 0x20} // "fmt "
	testRiffChunkFmt = []byte{
		0x10, 0x00, 0x00, 0x00, // LengthOfHeader
		0x01, 0x00, // AudioFormat
		0x01, 0x00, // NumOfChannels
		0x44, 0xac, 0x00, 0x00, // SampleRate
		0x88, 0x58, 0x01, 0x00, // BytesPerSec
		0x02, 0x00, // BytesPerBloc
		0x10, 0x00, // BitsPerSample
		0x64, 0x61, 0x74, 0x61, // "data"
	}

	wavWithOneSample []byte
)

func init() {
	var b bytes.Buffer
	b.Write(riff)
	b.Write([]byte{0x26, 0x00, 0x00, 0x00}) // chunkSize
	b.Write(wave)
	b.Write(fmt20)
	b.Write(testRiffChunkFmt)
	b.Write([]byte{0x02, 0x00, 0x00, 0x00}) // 2 bytes of samples - this part of the header is why wav length is capped at 32bit
	b.Write([]byte{0x01, 0x01})
	wavWithOneSample = b.Bytes()
}

func TestNewWavReader_inputTooLarge(t *testing.T) {
	t.Parallel()
	_, err := NewWavReader(
		bytes.NewReader([]byte{}),
		99999999999999999)
	assert.Equal(t, ErrInputToLarge, err)
}

func TestParseHeaders_complete(t *testing.T) {
	t.Parallel()
	// Parsing the header of an wav with 0 samples
	var b bytes.Buffer
	b.Write(riff)
	b.Write(chunkSize24)
	b.Write(wave)
	b.Write(fmt20)
	b.Write(testRiffChunkFmt)
	b.Write([]byte{0x00, 0x00, 0x00, 0x00})
	wavFile := bytes.NewReader(b.Bytes())
	wavReader, err := NewWavReader(wavFile, int64(b.Len()))
	assert.Nil(t, err)
	assert.Equal(t, 0, wavReader.GetSampleCount())
	assert.Equal(t, WavFile{
		SampleRate:      44100,
		Channels:        1,
		SignificantBits: 16,
	}, wavReader.GetWavFile())
}

func TestParseHeaders_tooShort(t *testing.T) {
	t.Parallel()
	wavData := append(riff, 0x08, 0x00)
	wavFile := bytes.NewReader(wavData)
	_, err := NewWavReader(wavFile, int64(len(wavData)))
	assert.NotNil(t, err)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
}

func TestParseHeaders_chunkFmtMissing(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	b.Write(riff)
	b.Write([]byte{0x04, 0x00, 0x00, 0x00}) // chunkSize
	b.Write(wave)
	wavFile := bytes.NewReader(b.Bytes())
	_, err := NewWavReader(wavFile, int64(b.Len()))
	assert.NotNil(t, err)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
}

func TestParseHeaders_chunkFmtTooShort(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	b.Write(riff)
	b.Write([]byte{0x08, 0x00, 0x00, 0x00}) // chunkSize
	b.Write(wave)
	b.Write(fmt20)
	wavFile := bytes.NewReader(b.Bytes())
	_, err := NewWavReader(wavFile, int64(b.Len()))
	assert.NotNil(t, err)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
}

func TestParseHeaders_chunkFmtTooShort2(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	b.Write(riff)
	b.Write([]byte{0x0a, 0x00, 0x00, 0x00}) // chunkSize
	b.Write(wave)
	b.Write(fmt20)
	b.Write([]byte{0, 0})
	wavFile := bytes.NewReader(b.Bytes())
	_, err := NewWavReader(wavFile, int64(b.Len()))
	assert.NotNil(t, err)
	assert.Equal(t, io.ErrUnexpectedEOF, err)
}

func TestParseHeaders_corruptRiff(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	b.Write([]byte{0x52, 0, 0x46, 0x46}) // "R\0FF"
	b.Write(chunkSize24)
	b.Write(wave)
	wavFile := bytes.NewReader(b.Bytes())
	_, err := NewWavReader(wavFile, int64(b.Len()))
	assert.NotNil(t, err)
	assert.Equal(t, ErrNotRiff, err)
}

func TestParseHeaders_chunkSizeNull(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	b.Write(riff)
	b.Write([]byte{0x00, 0x00, 0x00, 0x00}) // chunkSize
	b.Write(wave)
	b.Write(fmt20)
	wavFile := bytes.NewReader(b.Bytes())
	_, err := NewWavReader(wavFile, int64(b.Len()))
	assert.NotNil(t, err)
	assert.Equal(t, ErrIncorrectChunkSize{8, 16}, err, "bad error")
	assert.EqualError(t, err, "Incorrect ChunkSize. Got[8] Wanted[16]")
}

func TestParseHeaders_notWave(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	b.Write(riff)
	b.Write([]byte{0x09, 0x00, 0x00, 0x00}) // chunkSize
	b.Write([]byte{0x57, 0x42, 0x56, 0x45}) // "WBVE"
	b.Write(fmt20)
	b.Write([]byte{0})
	wavFile := bytes.NewReader(b.Bytes())
	_, err := NewWavReader(wavFile, int64(b.Len()))
	assert.NotNil(t, err)
	assert.Equal(t, ErrNotWave, err)
}

func TestParseHeaders_fmtNotSupported(t *testing.T) {
	t.Parallel()
	var b bytes.Buffer
	b.Write(riff)
	b.Write(chunkSize24)
	b.Write(wave)
	b.Write(fmt20)
	b.Write(testRiffChunkFmt)
	b.Write([]byte{0x00, 0x00, 0x00, 0x00})
	buf := b.Bytes()
	buf[21] = 2 // change byte 5 of riffChunk
	wavFile := bytes.NewReader(buf)
	_, err := NewWavReader(wavFile, int64(b.Len()))
	assert.NotNil(t, err)
	assert.Equal(t, ErrFormatNotSupported, err)
}

func TestReadSample_Raw(t *testing.T) {
	t.Parallel()
	wavFile := bytes.NewReader(wavWithOneSample)
	wavReader, err := NewWavReader(wavFile, int64(len(wavWithOneSample)))
	assert.Nil(t, err)
	assert.Equal(t, 1, wavReader.GetSampleCount())
	rawSample, err := wavReader.ReadRawSample()
	assert.Nil(t, err)
	assert.Equal(t, []byte{1, 1}, rawSample)
}

func TestReadSample(t *testing.T) {
	t.Parallel()
	wavFile := bytes.NewReader(wavWithOneSample)
	wavReader, err := NewWavReader(wavFile, int64(len(wavWithOneSample)))
	assert.Nil(t, err)
	assert.Equal(t, 1, wavReader.GetSampleCount())
	sample, err := wavReader.ReadSample()
	assert.Nil(t, err)
	assert.Equal(t, 257, sample)
}
