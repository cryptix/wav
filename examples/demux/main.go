package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/cryptix/wav"
)

// Demux a 2 channel audio file and save the left channel as a mono file
// http://blog.bjornroche.com/2013/05/the-abcs-of-pcm-uncompressed-digital.html

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: demux <file.wav>\n")
		os.Exit(1)
	}

	testInfo, err := os.Stat(os.Args[1])
	checkErr(err)

	testWav, err := os.Open(os.Args[1])
	checkErr(err)

	wavReader, err := wav.NewReader(testWav, testInfo.Size())
	checkErr(err)

	fileMeta := wavReader.GetFile()
	fmt.Println(dump(fileMeta))

	if fileMeta.Channels != 2 {
		log.Fatal("Please use a 2-channel audio file")
	}

	// A slice of many 1-4 byte samples
	var left [][]byte
	left = make([][]byte, fileMeta.NumberOfSamples/2)

	var i uint32
	var buf []byte
	for i = 0; i < fileMeta.NumberOfSamples/2; i++ {

		buf, err = wavReader.ReadRawSample()
		checkErr(err)
		left[i] = buf

		// Throw the right-channel away
		_, err = wavReader.ReadRawSample()
		checkErr(err)
	}

	filename := "mono-" + testInfo.Name()
	os.Remove(filename)

	f, err := os.Create(filename)
	checkErr(err)

	// Create the headers for our new mono file
	meta := wav.File{
		Channels:        1,
		SampleRate:      fileMeta.SampleRate,
		SignificantBits: fileMeta.SignificantBits,
	}

	writer, err := meta.NewWriter(f)
	checkErr(err)

	// Write to file
	for _, sample := range left {
		err = writer.WriteSample(sample)
		checkErr(err)
	}

	err = writer.Close()
	checkErr(err)

	fmt.Println("Created Mono File", f.Name())

	f, err = os.Open(f.Name())
	checkErr(err)

	stat, err := f.Stat()
	checkErr(err)

	wavReader, err = wav.NewReader(f, stat.Size())
	checkErr(err)

	fileMeta = wavReader.GetFile()
	fmt.Println(dump(fileMeta))
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func dump(v interface{}) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b)
}
