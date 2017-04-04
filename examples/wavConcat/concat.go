package main

import (
	"fmt"
	"github.com/rlfosterjr/wav"
	"os"
)

func concat(files []string, ofn string) {
	var wf = wav.File{
		Channels:        2,
		SampleRate:      44100,
		SignificantBits: 16,
	}

	of, _ := os.Create(ofn)
	defer of.Close()

	inc := len(files)
	ww, err := wf.NewWriter(of)

	if err != nil {
		fmt.Println(err)
	} else {
		for i := 0; i < inc; i++ {
			fmt.Println("File: " + files[i])
			f, _ := os.Open(files[i])
			stat, _ := f.Stat()
			wr, _ := wav.NewReader(f, stat.Size())
			fmt.Println(wr.String())

			var s uint32
			samps := wr.GetSampleCount()

			for s = 0; s < samps; s++ {
				cs, _ := wr.ReadRawSample()
				ww.WriteSample(cs)
			}
			f.Close()
		}
		ww.Close()
	}
}

//Test Wav Concat (stereo)
func main() {

	//concat incremnts
	inc := 3

	increments := make([]string, inc)
	increments[0] = "../../corpus/0009.wav"
	increments[1] = "../../corpus/0918.wav"
	increments[2] = "../../corpus/1827.wav"

	of, _ := os.Create("concat.wav")
	defer of.Close()

	var wf = wav.File{
		Channels:        2,
		SampleRate:      44100,
		SignificantBits: 16,
	}

	ww, err := wf.NewWriter(of)
	if err != nil {
		fmt.Println(err)
	} else {
		for i := 0; i < inc; i++ {
			fmt.Println("File: " + increments[i])
			f, _ := os.Open(increments[i])
			stat, _ := f.Stat()
			wr, _ := wav.NewReader(f, stat.Size())
			fmt.Println(wr.String())

			var s uint32
			samps := wr.GetSampleCount()

			for s = 0; s < samps; s++ {
				cs, _ := wr.ReadRawSample()
				ww.WriteSample(cs)
			}
			f.Close()
		}
		ww.Close()
	}
}
