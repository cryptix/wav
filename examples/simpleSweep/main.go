package main

import (
	"fmt"
	"math"
	"os"
	"time"

	"github.com/cryptix/wav"
)

const (
	bits = 32
	rate = 44100
)

func main() {
	wavOut, err := os.Create("Test.wav")
	checkErr(err)
	defer wavOut.Close()

	meta := wav.WavFile{
		Channels:        1,
		SampleRate:      rate,
		SignificantBits: bits,
	}

	writer, err := meta.NewWriter(wavOut)
	checkErr(err)
	defer writer.CloseFile()

	start := time.Now()

	var freq float64
	freq = 0.0001
	for n := 0; n < 50*rate; n += 1 {
		y := int32(0.8 * math.Pow(2, bits-1) * math.Sin(freq*float64(n)))
		freq += 0.000002

		err = writer.WriteInt32(y)
		checkErr(err)
	}

	fmt.Printf("Simulation Done. Took:%v\n", time.Since(start))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
