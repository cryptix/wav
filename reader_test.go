package wav

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/cheekybits/is"
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

func TestNewReader_inputTooLarge(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	_, err := NewReader(
		bytes.NewReader([]byte{}),
		99999999999999999)
	is.Equal(err, ErrInputToLarge)
}

func TestParseHeaders_complete(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	// Parsing the header of an wav with 0 samples
	var b bytes.Buffer
	b.Write(riff)
	b.Write(chunkSize24)
	b.Write(wave)
	b.Write(fmt20)
	b.Write(testRiffChunkFmt)
	b.Write([]byte{0x00, 0x00, 0x00, 0x00})
	wavFile := bytes.NewReader(b.Bytes())
	wavReader, err := NewReader(wavFile, int64(b.Len()))
	is.NoErr(err)
	is.Equal(uint32(0), wavReader.GetSampleCount())
	is.Equal(File{
		SampleRate:      44100,
		Channels:        1,
		SignificantBits: 16,
		AudioFormat:     1,
		Canonical:       true,
		BytesPerSecond:  88200,
	}, wavReader.GetFile())
}

func TestParseHeaders_tooShort(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	wavData := append(riff, 0x08, 0x00)
	wavFile := bytes.NewReader(wavData)
	_, err := NewReader(wavFile, int64(len(wavData)))
	is.Err(err)
	is.Equal(err, io.ErrUnexpectedEOF)
}

func TestParseHeaders_chunkFmtMissing(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	var b bytes.Buffer
	b.Write(riff)
	b.Write([]byte{0x04, 0x00, 0x00, 0x00}) // chunkSize
	b.Write(wave)
	wavFile := bytes.NewReader(b.Bytes())
	_, err := NewReader(wavFile, int64(b.Len()))
	is.Err(err)
	is.Equal(err, io.ErrUnexpectedEOF)
}

func TestParseHeaders_chunkFmtTooShort(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	var b bytes.Buffer
	b.Write(riff)
	b.Write([]byte{0x08, 0x00, 0x00, 0x00}) // chunkSize
	b.Write(wave)
	b.Write(fmt20)
	wavFile := bytes.NewReader(b.Bytes())
	_, err := NewReader(wavFile, int64(b.Len()))
	is.Err(err)
	is.Equal(err, io.ErrUnexpectedEOF)
}

func TestParseHeaders_chunkFmtTooShort2(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	var b bytes.Buffer
	b.Write(riff)
	b.Write([]byte{0x0a, 0x00, 0x00, 0x00}) // chunkSize
	b.Write(wave)
	b.Write(fmt20)
	b.Write([]byte{0, 0})
	wavFile := bytes.NewReader(b.Bytes())
	_, err := NewReader(wavFile, int64(b.Len()))
	is.Err(err)
	is.Equal(err, io.ErrUnexpectedEOF)
}

func TestParseHeaders_corruptRiff(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	var b bytes.Buffer
	b.Write([]byte{0x52, 0, 0x46, 0x46}) // "R\0FF"
	b.Write(chunkSize24)
	b.Write(wave)
	wavFile := bytes.NewReader(b.Bytes())
	_, err := NewReader(wavFile, int64(b.Len()))
	is.Err(err)
	is.Equal(ErrNotRiff, err)
}

func TestParseHeaders_chunkSizeNull(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	var b bytes.Buffer
	b.Write(riff)
	b.Write([]byte{0x00, 0x00, 0x00, 0x00}) // chunkSize
	b.Write(wave)
	b.Write(fmt20)
	wavFile := bytes.NewReader(b.Bytes())
	_, err := NewReader(wavFile, int64(b.Len()))
	is.Err(err)
	is.Equal(ErrIncorrectChunkSize{8, 16}, err)
	is.Equal("Incorrect ChunkSize. Got[8] Wanted[16]", err.Error())
}

func TestParseHeaders_notWave(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	var b bytes.Buffer
	b.Write(riff)
	b.Write([]byte{0x09, 0x00, 0x00, 0x00}) // chunkSize
	b.Write([]byte{0x57, 0x42, 0x56, 0x45}) // "WBVE"
	b.Write(fmt20)
	b.Write([]byte{0})
	wavFile := bytes.NewReader(b.Bytes())
	_, err := NewReader(wavFile, int64(b.Len()))
	is.Err(err)
	is.Equal(ErrNotWave, err)
}

func TestParseHeaders_fmtNotSupported(t *testing.T) {
	t.Parallel()
	is := is.New(t)
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
	_, err := NewReader(wavFile, int64(b.Len()))
	is.Err(err)
	is.Equal(ErrFormatNotSupported, err)
}

func TestReadSample_Raw(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	wavFile := bytes.NewReader(wavWithOneSample)
	wavReader, err := NewReader(wavFile, int64(len(wavWithOneSample)))
	is.NoErr(err)
	is.Equal(uint32(1), wavReader.GetSampleCount())
	rawSample, err := wavReader.ReadRawSample()
	is.NoErr(err)
	is.Equal([]byte{1, 1}, rawSample)
}

func TestReadSample(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	wavFile := bytes.NewReader(wavWithOneSample)
	wavReader, err := NewReader(wavFile, int64(len(wavWithOneSample)))
	is.NoErr(err)
	is.Equal(uint32(1), wavReader.GetSampleCount())
	sample, err := wavReader.ReadSample()
	is.NoErr(err)
	is.Equal(257, sample)
}

// panic: runtime error: invalid memory address or nil pointer dereference
// [signal 0xb code=0x1 addr=0x4 pc=0x4399fb]
//
// goroutine 1 [running]:
// github.com/cryptix/wav.(*Reader).parseHeaders(0xc208033720, 0x0, 0x0)
// 	/tmp/go-fuzz-build857960013/src/github.com/cryptix/wav/reader.go:191 +0xe3b
// github.com/cryptix/wav.NewReader(0x7f23a9550bd8, 0xc208037c80, 0x2d, 0xc208033720, 0x0, 0x0)
// 	/tmp/go-fuzz-build857960013/src/github.com/cryptix/wav/reader.go:64 +0x177
// github.com/cryptix/wav.Fuzz(0x7f23a92cf000, 0x2d, 0x100000, 0x2)
// 	/tmp/go-fuzz-build857960013/src/github.com/cryptix/wav/fuzz.go:12 +0x167
// github.com/dvyukov/go-fuzz/go-fuzz-dep.Main(0x570c60, 0x5d4200, 0x5f6, 0x5f6)
// 	/home/cryptix/go/src/github.com/dvyukov/go-fuzz/go-fuzz-dep/main.go:64 +0x309
// main.main()
// 	/tmp/go-fuzz-build857960013/src/go-fuzz-main/main.go:10 +0x4e
// exit status 2
func TestReadFuzzed_panic1(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	wavFile := strings.NewReader("RIFF%\x00\x00\x00WAVE0000\x10\x00\x00\x000000000000000000data00000")
	_, err := NewReader(wavFile, int64(wavFile.Len()))
	is.Err(err)
	is.Equal(ErrBrokenChunkFmt, err)
}

// panic: runtime error: integer divide by zero
// [signal 0x8 code=0x1 addr=0x439ae9 pc=0x439ae9]
//
// goroutine 1 [running]:
// github.com/cryptix/wav.(*Reader).parseHeaders(0xc208032cd0, 0x0, 0x0)
// 	/tmp/go-fuzz-build857960013/src/github.com/cryptix/wav/reader.go:200 +0xf29
// github.com/cryptix/wav.NewReader(0x7fbca32b6bd8, 0xc208037ef0, 0x2d, 0xc208032cd0, 0x0, 0x0)
// 	/tmp/go-fuzz-build857960013/src/github.com/cryptix/wav/reader.go:64 +0x177
// github.com/cryptix/wav.Fuzz(0x7fbca3035000, 0x2d, 0x100000, 0x2)
// 	/tmp/go-fuzz-build857960013/src/github.com/cryptix/wav/fuzz.go:12 +0x167
// github.com/dvyukov/go-fuzz/go-fuzz-dep.Main(0x570c60, 0x5d4200, 0x5f6, 0x5f6)
// 	/home/cryptix/go/src/github.com/dvyukov/go-fuzz/go-fuzz-dep/main.go:64 +0x309
// main.main()
// 	/tmp/go-fuzz-build857960013/src/go-fuzz-main/main.go:10 +0x4e
// exit status 2
func TestReadFuzzed_panic2(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	wavFile := strings.NewReader("RIFF%\x00\x00\x00WAVEfmt \x10\x00\x00\x00\x01\x00000000000000\a\x00data00000")
	_, err := NewReader(wavFile, int64(wavFile.Len()))
	is.Err(err)
	is.Equal(ErrNoBitsPerSample, err)
}
