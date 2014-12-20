package wav

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var wf = WavFile{
	SampleRate:      44100,
	Channels:        1,
	SignificantBits: 16,
}

func TestNewWavWriter(t *testing.T) {
	t.Parallel()
	f, err := ioutil.TempFile("", "wavPkgtest")
	assert.Nil(t, err)
	wr, err := wf.NewWriter(f)
	assert.Nil(t, err)
	assert.Nil(t, wr.Close())

	f, err = os.Open(f.Name())
	assert.Nil(t, err)

	b, err := ioutil.ReadAll(f)
	assert.Nil(t, err)
	assert.Len(t, b, 44)

	assert.Contains(t, string(b), string(riff))
	assert.Contains(t, string(b), string(wave))
	assert.Contains(t, string(b), string(fmt20))

	assert.Nil(t, os.Remove(f.Name()))
}

func bWriteByteSlice(sample []byte, samples int, b *testing.B) {
	b.StopTimer()
	f, err := ioutil.TempFile("", "wavPkgtest")
	assert.Nil(b, err)
	defer os.Remove(f.Name())
	wr, err := wf.NewWriter(f)
	defer wr.Close()
	assert.Nil(b, err)
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
	f, err := ioutil.TempFile("", "wavPkgtest")
	assert.Nil(b, err)
	defer os.Remove(f.Name())
	wr, err := wf.NewWriter(f)
	defer wr.Close()
	assert.Nil(b, err)
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
