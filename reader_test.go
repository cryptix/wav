package wav

import (
	"bytes"
	"io"

	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestNewWavReader(t *testing.T) {
	Convey("Parsing the header of an wav with 0 samples", t, func() {
		wavData := []byte{
			0x52, 0x49, 0x46, 0x46, // "RIFF"
			0x24, 0x00, 0x00, 0x00, // chunkSize
			0x57, 0x41, 0x56, 0x45, // "WAVE"
			0x66, 0x6d, 0x74, 0x20, // "fmt "

			// riffChunkFmt
			0x10, 0x00, 0x00, 0x00, // LengthOfHeader
			0x01, 0x00, // AudioFormat
			0x01, 0x00, // NumOfChannels
			0x44, 0xac, 0x00, 0x00, // SampleRate
			0x88, 0x58, 0x01, 0x00, // BytesPerSec
			0x02, 0x00, // BytesPerBloc
			0x10, 0x00, // BitsPerSample

			0x64, 0x61, 0x74, 0x61, // "data"
			0x00, 0x00, 0x00, 0x00,
		}
		wavFile := bytes.NewReader(wavData)

		wavReader, err := NewWavReader(wavFile, int64(len(wavData)))
		So(err, ShouldBeNil)

		Convey("GetSampleCount should return 0", func() {
			So(wavReader.GetSampleCount(), ShouldEqual, 0)
		})

		Convey("GetWavFile should return the correct information", func() {
			So(wavReader.GetWavFile(), ShouldResemble, WavFile{
				SampleRate:      44100,
				Channels:        1,
				SignificantBits: 16,
			})
		})
	})

	Convey("Refusing to parse enourmes wav files - returns ErrInputToLarge", t, func() {
		wavData := []byte{}
		wavFile := bytes.NewReader(wavData)

		_, err := NewWavReader(wavFile, 99999999999999999)
		So(err, ShouldEqual, ErrInputToLarge)
	})
}

func TestParseHeaders(t *testing.T) {
	Convey("Returning io.ErrUnexpectedEOF when the headers are too short", t, func() {

		Convey("when the RIFFheader is too short", func() {
			wavData := []byte{
				0x52, 0x49, 0x46, 0x46, // "RIFF"
				0x08, 0x00,
			}
			wavFile := bytes.NewReader(wavData)

			reader, err := NewWavReader(wavFile, int64(len(wavData)))
			So(reader, ShouldBeNil)
			So(err, ShouldResemble, io.ErrUnexpectedEOF)
		})

		Convey("when the chunkFmt is missing", func() {
			wavData := []byte{
				0x52, 0x49, 0x46, 0x46, // "RIFF"
				0x04, 0x00, 0x00, 0x00, // chunkSize
				0x57, 0x41, 0x56, 0x45, // "WAVE"
			}
			wavFile := bytes.NewReader(wavData)

			reader, err := NewWavReader(wavFile, int64(len(wavData)))
			So(reader, ShouldBeNil)
			So(err, ShouldResemble, io.ErrUnexpectedEOF)
		})

		Convey("when the chunkFmt is too short", func() {
			wavData := []byte{
				0x52, 0x49, 0x46, 0x46, // "RIFF"
				0x08, 0x00, 0x00, 0x00, // chunkSize
				0x57, 0x41, 0x56, 0x45, // "WAVE"
				0x66, 0x6d, 0x74, 0x20, // "fmt "
			}
			wavFile := bytes.NewReader(wavData)

			reader, err := NewWavReader(wavFile, int64(len(wavData)))
			So(reader, ShouldBeNil)
			So(err, ShouldResemble, io.ErrUnexpectedEOF)
		})

		Convey("when the chunkFmt is too short", func() {
			wavData := []byte{
				0x52, 0x49, 0x46, 0x46, // "RIFF"
				0x0a, 0x00, 0x00, 0x00, // chunkSize
				0x57, 0x41, 0x56, 0x45, // "WAVE"
				0x66, 0x6d, 0x74, 0x20, // "fmt "
				0, 0,
			}
			wavFile := bytes.NewReader(wavData)

			reader, err := NewWavReader(wavFile, int64(len(wavData)))
			So(reader, ShouldBeNil)
			So(err, ShouldResemble, io.ErrUnexpectedEOF)
		})
	})

	Convey("Parsing the a corrupted RIFF header returns ErrNotRiff", t, func() {
		wavData := []byte{
			0x52, 0, 0x46, 0x46, // "R\0FF"
			0x24, 0x00, 0x00, 0x00, // chunkSize
			0x57, 0x41, 0x56, 0x45, // "WAVE"
		}
		wavFile := bytes.NewReader(wavData)

		reader, err := NewWavReader(wavFile, int64(len(wavData)))
		So(reader, ShouldBeNil)
		So(err, ShouldEqual, ErrNotRiff)
	})

	Convey("Parsing an incorrect chunkSize returns ErrIncorrectChunkSize", t, func() {
		wavData := []byte{
			0x52, 0x49, 0x46, 0x46, // "RIFF"
			0x00, 0x00, 0x00, 0x00, // chunkSize=0
			0x57, 0x41, 0x56, 0x45, // "WAVE"
			0x66, 0x6d, 0x74, 0x20, // "fmt "
		}
		wavFile := bytes.NewReader(wavData)

		reader, err := NewWavReader(wavFile, int64(len(wavData)))
		So(reader, ShouldBeNil)
		So(err, ShouldResemble, ErrIncorrectChunkSize{8, 16})
		So(err.Error(), ShouldEqual, "Incorrect ChunkSize. Got[8] Wanted[16]")
	})

	Convey("Parsing an incorrect WAVE token returns ErrNotWave", t, func() {
		wavData := []byte{
			0x52, 0x49, 0x46, 0x46, // "RIFF"
			0x09, 0x00, 0x00, 0x00, // chunkSize
			0x57, 0x42, 0x56, 0x45, // "WBVE"
			0x66, 0x6d, 0x74, 0x20, // "fmt "
			0,
		}
		wavFile := bytes.NewReader(wavData)

		reader, err := NewWavReader(wavFile, int64(len(wavData)))
		So(reader, ShouldBeNil)
		So(err, ShouldEqual, ErrNotWave)
	})

	Convey("Only uncompressed PCM is supported - ErrFormatNotSupported", t, func() {
		wavData := []byte{
			0x52, 0x49, 0x46, 0x46, // "RIFF"
			0x24, 0x00, 0x00, 0x00, // chunkSize
			0x57, 0x41, 0x56, 0x45, // "WAVE"
			0x66, 0x6d, 0x74, 0x20, // "fmt "

			// riffChunkFmt
			0x10, 0x00, 0x00, 0x00, // LengthOfHeader
			0x02, 0x00, // AudioFormat
			0x01, 0x00, // NumOfChannels
			0x44, 0xac, 0x00, 0x00, // SampleRate
			0x88, 0x58, 0x01, 0x00, // BytesPerSec
			0x02, 0x00, // BytesPerBloc
			0x10, 0x00, // BitsPerSample

			0x64, 0x61, 0x74, 0x61, // "data"
			0x00, 0x00, 0x00, 0x00,
		}
		wavFile := bytes.NewReader(wavData)

		reader, err := NewWavReader(wavFile, int64(len(wavData)))
		So(reader, ShouldBeNil)
		So(err, ShouldEqual, ErrFormatNotSupported)
	})
}

func TestReadRawSample(t *testing.T) {
	var (
		err       error
		wavReader *WavReader
	)

	Convey("Parsing the header of an wav with 0 samples", t, func() {
		wavData := []byte{
			0x52, 0x49, 0x46, 0x46, // "RIFF"
			0x26, 0x00, 0x00, 0x00, // chunkSize
			0x57, 0x41, 0x56, 0x45, // "WAVE"
			0x66, 0x6d, 0x74, 0x20, // "fmt "

			// riffChunkFmt
			0x10, 0x00, 0x00, 0x00, // LengthOfHeader
			0x01, 0x00, // AudioFormat
			0x01, 0x00, // NumOfChannels
			0x44, 0xac, 0x00, 0x00, // SampleRate
			0x88, 0x58, 0x01, 0x00, // BytesPerSec
			0x02, 0x00, // BytesPerBloc
			0x10, 0x00, // BitsPerSample

			0x64, 0x61, 0x74, 0x61, // "data"
			0x02, 0x00, 0x00, 0x00,
			0x01, 0x01,
		}
		wavFile := bytes.NewReader(wavData)

		wavReader, err = NewWavReader(wavFile, int64(len(wavData)))
		So(err, ShouldBeNil)

		Convey("GetSampleCount should return 0", func() {
			So(wavReader.GetSampleCount(), ShouldEqual, 1)
		})

		Convey("ReadRawSample should match", func() {
			sample, err := wavReader.ReadRawSample()
			So(err, ShouldBeNil)
			So(sample, ShouldResemble, []byte{1, 1})
		})

		Convey("ReadSample should match", func() {
			sample, err := wavReader.ReadSample()
			So(err, ShouldBeNil)
			So(sample, ShouldEqual, 257)
		})
	})
}
