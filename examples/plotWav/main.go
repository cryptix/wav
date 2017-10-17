package main

import (
	"fmt"
	"io"
	"os"
	"path"
	"strings"

	"github.com/cryptix/wav"

	"gonum.org/v1/plot"
	"gonum.org/v1/plot/plotter"
	"gonum.org/v1/plot/plotutil"
	"gonum.org/v1/plot/vg"
)

func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "Usage: plotWav <file.wav>\n")
		os.Exit(1)
	}

	// open file
	testInfo, err := os.Stat(os.Args[1])
	checkErr(err)

	testWav, err := os.Open(os.Args[1])
	checkErr(err)

	wavReader, err := wav.NewReader(testWav, testInfo.Size())
	checkErr(err)

	// File informations
	fmt.Println(wavReader)

	// limit sample count
	sampleCnt := wavReader.GetSampleCount()
	if sampleCnt > 10000 {
		sampleCnt = 10000
	}

	// setup plotter
	p, err := plot.New()
	checkErr(err)

	p.Title.Text = "Waveplot"
	p.X.Label.Text = "t"
	p.Y.Label.Text = "Ampl"

	pts := make(plotter.XYs, sampleCnt)

	// read samples and construct points for plot
	for i := range pts {
		n, err := wavReader.ReadSample()
		if err == io.EOF {
			break
		}
		checkErr(err)

		pts[i].X = float64(i)
		pts[i].Y = float64(n)
	}

	err = plotutil.AddLinePoints(p, "", pts)
	checkErr(err)

	// construct output filename
	inputFname := path.Base(os.Args[1])
	plotFname := strings.Split(inputFname, ".")[0] + ".pdf"

	if err := p.Save(10*vg.Inch, 4*vg.Inch, plotFname); err != nil {
		panic(err)
	}
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}
