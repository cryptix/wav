package wav

import (
	"bytes"

	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestWavReader(t *testing.T) {
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

		Convey("GetWavFile should return", func() {
			So(wavReader.GetWavFile(), ShouldResemble, WavFile{
				SampleRate:      44100,
				Channels:        1,
				SignificantBits: 16,
			})
		})
	})
}
