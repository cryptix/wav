package main

import (
	"fmt"
	"log"
	"os"

	"github.com/xeoncross/wav"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: simpleReadEvery <file.wav>\n")
		os.Exit(1)
	}
	testInfo, err := os.Stat(os.Args[1])
	checkErr(err)

	testWav, err := os.Open(os.Args[1])
	checkErr(err)

	wavReader, err := wav.NewReader(testWav, testInfo.Size())
	checkErr(err)

	fmt.Println("Hello, wav")
	fmt.Println(wavReader)

	// Load file meta
	var meta wav.File
	meta = wavReader.GetFile()

	// Every half a second read a sample
	readSampleRate := meta.SampleRate / uint32(2)
	fmt.Println("Read a sample every", readSampleRate)

	samples, err := wavReader.ReadSampleEvery(readSampleRate)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Samples found %d, Estimated: %d\n", len(samples), meta.NumberOfSamples/readSampleRate+1)

	var second uint32
	for i, sample := range samples {

		pos := uint32(i)

		// What second of audio are we on?
		second = readSampleRate * pos / meta.SampleRate

		fmt.Printf("Second %d\tSample: %d\tAmplitude: %d\n", second, pos*readSampleRate, sample)
	}

}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
