package wav

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type WavWriter struct {
	output  io.WriteSeeker
	options WavFile

	samples bytes.Buffer

	samplesWritten int32
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

	// leave space for the data sie
	_, err = wr.output.Seek(4, os.SEEK_CUR)
	if err != nil {
		return
	}

	return
}

func (w *WavWriter) WriteSample(sample []byte) error {
	if len(sample)*8 != int(w.options.SignificantBits) {
		return fmt.Errorf("Incorrect Sample Length %d", len(sample))
	}

	n, err := w.samples.Write(sample)
	if err != nil {
		return err
	}
	w.samplesWritten += 1
	w.bytesWritten += n

	return nil
}

func (w *WavWriter) CloseFile() error {
	_, err := w.output.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	header := riffHeader{
		ChunkSize: uint32(w.bytesWritten + 4),
	}
	copy(header.Ftype[:], tokenRiff[:])
	copy(header.ChunkFormat[:], tokenWaveFormat[:])

	err = binary.Write(w.output, binary.LittleEndian, header)
	if err != nil {
		return err
	}

	// write data chunk size
	_, err = w.output.Seek(0x28, os.SEEK_SET)
	if err != nil {
		return err
	}

	// write chunk size
	var dataSize int32
	dataSize = w.samplesWritten * 4
	err = binary.Write(w.output, binary.LittleEndian, dataSize)
	if err != nil {
		return err
	}

	// there must be a cleaner way to do this..
	// then we could cleanup the whole header writing, too
	sampleReader := bytes.NewReader(w.samples.Bytes())
	n, err := io.Copy(w.output, sampleReader)
	if n != int64(w.samples.Len()) {
		panic("did nod copy enough.. WTF?")
	}

	return nil
}
