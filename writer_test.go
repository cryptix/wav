package wav

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewWavWriter(t *testing.T) {
	t.Parallel()
	f, err := ioutil.TempFile("", "wavPkgtest")
	assert.Nil(t, err)
	wf := WavFile{
		SampleRate:      44100,
		Channels:        1,
		SignificantBits: 16,
	}
	wr, err := wf.NewWriter(f)
	assert.Nil(t, err)
	assert.Nil(t, wr.Close())

	_, err = f.Seek(0, os.SEEK_SET)
	assert.Nil(t, err)

	b, err := ioutil.ReadAll(f)
	assert.Nil(t, err)
	assert.Len(t, b, 44)
	assert.Contains(t, string(b), string(riff))
	assert.Contains(t, string(b), string(wave))
	assert.Contains(t, string(b), string(fmt20))

	assert.Nil(t, os.Remove(f.Name()))
}

func benchmarkSilence(samples int, b *testing.B) {
	b.StopTimer()
	f, err := ioutil.TempFile("", "wavPkgtest")
	assert.Nil(b, err)
	defer os.Remove(f.Name())
	wf := WavFile{
		SampleRate:      44100,
		Channels:        1,
		SignificantBits: 16,
	}
	wr, err := wf.NewWriter(f)
	defer wr.Close()
	assert.Nil(b, err)

	b.SetBytes(int64(2 * samples))

	b.StartTimer()
	for i := 0; i < b.N; i++ {

		// 1/10 second
		for i := 0; i < samples; i++ {

			if err := wr.WriteSample([]byte{0, 0}); err != nil {
				b.Fatal(err)
			}
		}

	}
}

func BenchmarkWriteSilence16sample(b *testing.B) { benchmarkSilence(16, b) }
func BenchmarkWriteSilence32sample(b *testing.B) { benchmarkSilence(32, b) }
func BenchmarkWriteSilence64sample(b *testing.B) { benchmarkSilence(64, b) }
func BenchmarkWriteSilence10thSec(b *testing.B)  { benchmarkSilence(44100/10, b) }
func BenchmarkWriteSilenceHalfSec(b *testing.B)  { benchmarkSilence(44100/2, b) }
func BenchmarkWriteSilenceOneSec(b *testing.B)   { benchmarkSilence(44100, b) }
