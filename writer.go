package wav

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"io"
	"os"
)

type output interface {
	io.Writer
	io.Seeker
	io.Closer
}

// Writer encapsulates a io.WriteSeeker and supplies Functions for writing samples
type Writer struct {
	output
	options   File
	sampleBuf *bufio.Writer

	samplesWritten int32
	bytesWritten   int
}

// NewWriter creates a new WaveWriter and writes the header to it
func (file File) NewWriter(out output) (wr *Writer, err error) {
	if file.Channels != 1 {
		err = fmt.Errorf("sorry, only mono currently")
		return
	}

	wr = &Writer{}
	wr.output = out
	wr.sampleBuf = bufio.NewWriter(out)
	wr.options = file

	// write header when close to get correct number of samples
	wr.samplesWritten = 0
	_, err = wr.Seek(12, os.SEEK_SET)
	if err != nil {
		return
	}

	// fmt.Fprintf(wr, "%s", tokenChunkFmt)
	n, err := wr.Write(tokenChunkFmt[:])
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

	err = binary.Write(wr, binary.LittleEndian, chunkFmt)
	if err != nil {
		return
	}
	wr.bytesWritten += 20 //sizeof riffChunkFmt

	n, err = wr.Write(tokenData[:])
	if err != nil {
		return
	}
	wr.bytesWritten += n

	// leave space for the data size
	_, err = wr.Seek(4, os.SEEK_CUR)
	if err != nil {
		return
	}

	return
}

// WriteInt32 writes the sample to the file using the binary package
func (w *Writer) WriteInt32(sample int32) error {
	err := binary.Write(w.sampleBuf, binary.LittleEndian, sample)
	if err != nil {
		return err
	}

	w.samplesWritten++
	w.bytesWritten += 4

	return err
}

// WriteSample writes a []byte array to file without conversion
func (w *Writer) WriteSample(sample []byte) error {
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
func (w *Writer) GetDumbWriter() (wr output, countPtr *int32, err error) {
	if w.samplesWritten != 0 {
		return nil, nil, fmt.Errorf("Please only use this on its own")
	}

	return w, &w.samplesWritten, nil
}

// Close corrects the filesize information in the header
func (w *Writer) Close() error {

	if err := w.sampleBuf.Flush(); err != nil {
		return err
	}

	_, err := w.Seek(0, os.SEEK_SET)
	if err != nil {
		return err
	}

	header := riffHeader{
		ChunkSize: uint32(w.bytesWritten + 8),
	}
	copy(header.Ftype[:], tokenRiff[:])
	copy(header.ChunkFormat[:], tokenWaveFormat[:])

	err = binary.Write(w, binary.LittleEndian, header)
	if err != nil {
		return err
	}

	// write data chunk size
	_, err = w.Seek(0x28, os.SEEK_SET)
	if err != nil {
		return err
	}

	// write chunk size
	err = binary.Write(w, binary.LittleEndian, int32(w.bytesWritten))
	if err != nil {
		return err
	}

	return w.output.Close()
}
