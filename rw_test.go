package wav

import (
	"io/ioutil"
	"math"
	"os"
	"testing"

	"github.com/cheekybits/is"
)

func TestWriteRead_Int32(t *testing.T) {
	is := is.New(t)

	const (
		bits = 32
		rate = 44100
	)

	f, err := ioutil.TempFile("", "wavPkgtest")
	is.NoErr(err)

	testFname := f.Name()

	meta := File{
		Channels:        1,
		SampleRate:      rate,
		SignificantBits: bits,
	}

	writer, err := meta.NewWriter(f)
	is.NoErr(err)

	var freq float64
	freq = 0.0001

	// one second
	for n := 0; n < rate; n += 1 {
		y := int32(0.8 * math.Pow(2, bits-1) * math.Sin(freq*float64(n)))
		freq += 0.000002

		err = writer.WriteInt32(y)
		if err != nil {
			t.Fatal(err)
		}
	}
	err = writer.Close()
	is.NoErr(err)

	f, err = os.Open(testFname)
	is.NoErr(err)

	stat, err := f.Stat()
	is.NoErr(err)

	_, err = NewReader(f, stat.Size())
	is.NoErr(err)

	is.NoErr(os.Remove(testFname))
}

func TestWriteRead_Sample(t *testing.T) {
	is := is.New(t)

	const (
		bits = 16
		rate = 44100
	)

	f, err := ioutil.TempFile("", "wavPkgtest")
	is.NoErr(err)

	testFname := f.Name()

	meta := File{
		Channels:        1,
		SampleRate:      rate,
		SignificantBits: bits,
	}

	writer, err := meta.NewWriter(f)
	is.NoErr(err)

	var freq float64
	freq = 0.0001

	var s = make([]byte, 2)
	// one second
	for n := 0; n < rate; n += 1 {
		y := int32(0.8 * math.Pow(2, bits-1) * math.Sin(freq*float64(n)))
		freq += 0.000002

		// s[3] = byte((y >> 24) & 0xFF)
		// s[2] = byte((y >> 16) & 0xFF)
		s[1] = byte((y >> 8) & 0xFF)
		s[0] = byte(y & 0xFF)

		err = writer.WriteSample(s)
		if err != nil {
			t.Fatal(err)
		}
	}
	err = writer.Close()
	is.NoErr(err)

	f, err = os.Open(testFname)
	is.NoErr(err)

	stat, err := f.Stat()
	is.NoErr(err)

	_, err = NewReader(f, stat.Size())
	is.NoErr(err)

	is.NoErr(os.Remove(testFname))
}
