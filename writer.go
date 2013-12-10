package wav

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type WavWriter struct {
	output  io.WriteSeeker
	options WavFile

	samplesWritten int
	bytesWritten   int
}

func (file WavFile) NewWriter(out io.WriteSeeker) (wr *WavWriter, err error) {
	if file.Channels != 1 {
		err = fmt.Errorf("Sorry, only mono currently")
		return
	}

	wr = &WavWriter{}
	wr.output = out
	wr.options = file

	// write header when close to get correct number of samples
	wr.samplesWritten = 0
	wr.output.Seek(12, os.SEEK_SET)

	// fmt.Fprintf(wr.output, "%s", tokenChunkFmt)
	n, err := wr.output.Write(tokenChunkFmt[:])
	if err != nil {
		return
	}
	wr.bytesWritten += n

	chunkFmt := riffChunkFmt{
		LengthOfHeader: 16,
		AudioFormat:    1,
		NumChannels:    file.Channels,
		SampleRate:     file.SampleRate,
		BytesPerSec:    uint32(file.Channels) * file.SampleRate * uint32(file.SignificantBits) / 8,
		BytesPerBloc:   file.SignificantBits / 8 * file.Channels,
		BitsPerSample:  file.SignificantBits,
	}

	err = binary.Write(wr.output, binary.LittleEndian, chunkFmt)
	if err != nil {
		return
	}
	wr.bytesWritten += 20 //sizeof riffChunkFmt

	n, err = wr.output.Write(tokenData[:])
	if err != nil {
		return
	}
	wr.bytesWritten += n

	return
}

func (w *WavWriter) WriteSample(sample []byte) error {
	if len(sample)*8 != int(w.options.SignificantBits) {
		return fmt.Errorf("Incorrect Sample Length %d", len(sample))
	}

	binary.Write(w.output, binary.LittleEndian, sample)
	w.samplesWritten += 1
	w.bytesWritten += int(w.options.SignificantBits) / 8

	return nil
}

func (w *WavWriter) CloseFile() error {
	_, err := w.output.Seek(0, os.SEEK_SET)

	header := riffHeader{
		ChunkSize: uint32(w.bytesWritten + 4),
	}
	copy(header.Ftype[:], tokenRiff[:])
	copy(header.ChunkFormat[:], tokenWaveFormat[:])

	err = binary.Write(w.output, binary.LittleEndian, header)
	if err != nil {
		return err
	}

	return nil
}
