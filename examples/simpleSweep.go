package main

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"github.com/cryptix/wav"
	"math"
	"os"
	"time"
)

const (
	bits = 16
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

	// sampleBuf := make([]byte, bits/8)
	var sampleBuf bytes.Buffer

	start := time.Now()

	var freq float64
	freq = 0.0001
	for n := 0; n < 0*rate; n += 1 {
		y := int16(0.9 * math.Pow(2, bits-1) * math.Sin(freq*float64(n)))
		freq += 0.000001

		sampleBuf.Reset()
		binary.Write(&sampleBuf, binary.LittleEndian, y)

		err = writer.WriteSample(sampleBuf.Bytes())
		checkErr(err)
	}

	fmt.Printf("Simulation Done. Took:%v\n", time.Since(start))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
