package main

import (
	"fmt"
	"io"
	"os"

	"github.com/cryptix/wav"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: simpleRead <file.wav>\n")
		os.Exit(1)
	}
	testInfo, err := os.Stat(os.Args[1])
	checkErr(err)

	testWav, err := os.Open(os.Args[1])
	checkErr(err)

	wavReader, err := wav.NewWavReader(testWav, testInfo.Size())
	checkErr(err)

	fmt.Println("Hello, wav")
	fmt.Println(wavReader)

sampleLoop:
	for {
		s, err := wavReader.ReadRawSample()
		if err == io.EOF {
			break sampleLoop
		} else if err != nil {
			panic(err)
		}

		fmt.Printf("Sample: <%v>\n", s)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
