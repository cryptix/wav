package wav

import (
	"encoding/binary"
	"fmt"
	"io"
)

type WavWriter struct {
	output io.WriteSeeker
}

func NewWavWriter(out io.WriteSeeker, channels, samplefreq, bps int) (wr *WavWriter, err error) {
	if channels != 1 {
		err = fmt.Errorf("Sorry, only mono currently")
		return
	}

	wr = &WavWriter{}
	wr.output = out

	// todo set size at last..?
	size := 28
	header := riffHeader{
		ChunkSize: uint32(size),
	}
	copy(header.Ftype[:], WAVriffType[:])
	copy(header.ChunkFormat[:], WAVchunkFormat[:])

	err = binary.Write(wr.output, binary.LittleEndian, header)
	if err != nil {
		return
	}

	chunkFmt := riffChunkFmt{
		AudioFormat:   1,
		NumChannels:   uint16(channels),
		SampleFreq:    uint32(samplefreq),
		BitsPerSample: uint16(bps),
	}

	fmt.Fprintf(wr.output, "%s", WAVTokenFmt)

	err = binary.Write(wr.output, binary.LittleEndian, chunkFmt)
	if err != nil {
		return
	}

	fmt.Fprintf(wr.output, "%s", WAVTokenData)

	return
}
