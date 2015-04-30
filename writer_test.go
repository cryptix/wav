package wav

import (
	"bytes"
	"io/ioutil"
	"os"
	"testing"

	"github.com/cheekybits/is"
)

var wf = File{
	SampleRate:      44100,
	Channels:        1,
	SignificantBits: 16,
}

func TestNewWriter_Header(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	f, err := ioutil.TempFile("", "wavPkgtest")
	is.NoErr(err)
	wr, err := wf.NewWriter(f)
	is.NoErr(err)
	is.Nil(wr.Close())

	f, err = os.Open(f.Name())
	is.NoErr(err)

	b, err := ioutil.ReadAll(f)
	is.NoErr(err)
	is.Equal(len(b), 44)

	is.True(bytes.Contains(b, riff))
	is.True(bytes.Contains(b, wave))
	is.True(bytes.Contains(b, fmt20))

	is.Nil(os.Remove(f.Name()))
}

func TestNewWriter_1Sample(t *testing.T) {
	t.Parallel()
	is := is.New(t)
	f, err := ioutil.TempFile("", "wavPkgtest")
	is.NoErr(err)
	wr, err := wf.NewWriter(f)
	is.NoErr(err)

	err = wr.WriteSample([]byte{1, 1})
	is.NoErr(err)

	is.Nil(wr.Close())

	f, err = os.Open(f.Name())
	is.NoErr(err)

	b, err := ioutil.ReadAll(f)
	is.NoErr(err)
	is.Equal(len(b), 46)

	is.True(bytes.Contains(b, riff))
	is.True(bytes.Contains(b, wave))
	is.True(bytes.Contains(b, fmt20))

	is.Nil(os.Remove(f.Name()))
}

func bWriteByteSlice(sample []byte, samples int, b *testing.B) {
	b.StopTimer()
	is := is.New(b)
	f, err := ioutil.TempFile("", "wavPkgtest")
	is.NoErr(err)
	defer os.Remove(f.Name())
	wr, err := wf.NewWriter(f)
	defer wr.Close()
	is.NoErr(err)
	b.SetBytes(int64(2 * samples))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < samples; i++ {
			if err := wr.WriteSample(sample); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkWriteBuf_16sample(b *testing.B) { bWriteByteSlice([]byte{0, 0}, 16, b) }
func BenchmarkWriteBuf_32sample(b *testing.B) { bWriteByteSlice([]byte{0, 0}, 32, b) }
func BenchmarkWriteBuf_64sample(b *testing.B) { bWriteByteSlice([]byte{0, 0}, 64, b) }
func BenchmarkWriteBuf_10thSec(b *testing.B)  { bWriteByteSlice([]byte{0, 0}, 44100/10, b) }
func BenchmarkWriteBuf_HalfSec(b *testing.B)  { bWriteByteSlice([]byte{0, 0}, 44100/2, b) }
func BenchmarkWriteBuf_1Sec(b *testing.B)     { bWriteByteSlice([]byte{0, 0}, 44100, b) }
func BenchmarkWriteBuf_2Sec(b *testing.B)     { bWriteByteSlice([]byte{0, 0}, 2*44100, b) }

func benchWriteInt(sample int32, samples int, b *testing.B) {
	b.StopTimer()
	is := is.New(b)
	f, err := ioutil.TempFile("", "wavPkgtest")
	is.NoErr(err)
	defer os.Remove(f.Name())
	wr, err := wf.NewWriter(f)
	defer wr.Close()
	is.NoErr(err)
	b.SetBytes(int64(2 * samples))
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		for i := 0; i < samples/2; i++ {
			if err := wr.WriteInt32(sample); err != nil {
				b.Fatal(err)
			}
		}
	}
}

func BenchmarkWriteInt32_16sample(b *testing.B) { benchWriteInt(0, 16, b) }
func BenchmarkWriteInt32_32sample(b *testing.B) { benchWriteInt(0, 32, b) }
func BenchmarkWriteInt32_64sample(b *testing.B) { benchWriteInt(0, 64, b) }
func BenchmarkWriteInt32_10thSec(b *testing.B)  { benchWriteInt(0, 44100/10, b) }
func BenchmarkWriteInt32_HalfSec(b *testing.B)  { benchWriteInt(0, 44100/2, b) }
func BenchmarkWriteInt32_1Sec(b *testing.B)     { benchWriteInt(0, 44100, b) }
func BenchmarkWriteInt32_2Sec(b *testing.B)     { benchWriteInt(0, 2*44100, b) }
