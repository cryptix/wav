package wav

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

// WavWriter encapsulates a io.WriteSeeker and supplies Functions for writing samples
type WavWriter struct {
	output    io.WriteSeeker
	options   WavFile
	sampleBuf *bufio.Writer

	samplesWritten int32
	bytesWritten   int
}

// NewWriter creates a new WaveWriter and writes the header to it
func (file WavFile) NewWriter(out io.WriteSeeker) (wr *WavWriter, err error) {
	if file.Channels != 1 {
		err = fmt.Errorf("sorry, only mono currently")
		return
	}

	wr = &WavWriter{}
	wr.output = out
	wr.sampleBuf = bufio.NewWriter(out)
	wr.options = file

	// write header when close to get correct number of samples
	wr.samplesWritten = 0
	_, err = wr.output.Seek(12, os.SEEK_SET)
	if err != nil {
		return
	}

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

	// leave space for the data size
	_, err = wr.output.Seek(4, os.SEEK_CUR)
	if err != nil {
		return
	}

	return
}

// WriteInt32 writes the sample to the file using the binary package
func (w *WavWriter) WriteInt32(sample int32) error {
	err := binary.Write(w.sampleBuf, binary.LittleEndian, sample)
	if err != nil {
		return err
	}

	w.samplesWritten++
	w.bytesWritten += 4

	return err
}

// WriteSample writes a []byte array to file without conversion
func (w *WavWriter) WriteSample(sample []byte) error {
	if len(sample)*8 != int(w.options.SignificantBits) {
		return fmt.Errorf("incorrect Sample Length %d", len(sample))
	}

	n, err := w.sampleBuf.Write(sample)
	if err != nil {
		return err
	}

	w.samplesWritten++
	w.bytesWritten += n

	return nil
}

// GetDumbWriter gives you a std io.Writer, starting from the first sample. usefull for piping data.
func (w *WavWriter) GetDumbWriter() (wr io.Writer, countPtr *int32, err error) {
	if w.samplesWritten != 0 {
		return nil, nil, fmt.Errorf("Please only use this on its own")
	}

	return w.output, &w.samplesWritten, nil
}

// CloseFile corrects the filesize information in the header
func (w *WavWriter) CloseFile() error {
	_, err := w.output.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	header := riffHeader{
		ChunkSize: uint32(w.bytesWritten + 8),
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

	return nil
}
