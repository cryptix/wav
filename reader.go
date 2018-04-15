package wav

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"sort"
	"time"
)

// Reader wraps WAV stream
type Reader struct {
	input io.ReadSeeker
	size  int64

	header   *riffHeader
	chunkFmt *riffChunkFmt

	canonical      bool
	extraChunk     bool
	firstSamplePos uint32
	dataBlocSize   uint32
	bytesPerSample uint32
	duration       time.Duration

	samplesRead uint32
	numSamples  uint32
}

func (wav Reader) String() string {
	msg := fmt.Sprintln("File informations")
	msg += fmt.Sprintln("=================")
	msg += fmt.Sprintf("File size         : %d bytes\n", wav.size)
	msg += fmt.Sprintf("Canonical format  : %v\n", wav.canonical && !wav.extraChunk)
	// chunk fmt
	msg += fmt.Sprintf("Audio format      : %d\n", wav.chunkFmt.AudioFormat)
	msg += fmt.Sprintf("Number of channels: %d\n", wav.chunkFmt.NumChannels)
	msg += fmt.Sprintf("Sampling rate     : %d Hz\n", wav.chunkFmt.SampleRate)
	msg += fmt.Sprintf("Sample size       : %d bits\n", wav.chunkFmt.BitsPerSample)
	// calculated
	msg += fmt.Sprintf("Number of samples : %d\n", wav.numSamples)
	msg += fmt.Sprintf("Sound size        : %d bytes\n", wav.dataBlocSize)
	msg += fmt.Sprintf("Sound duration    : %v\n", wav.duration)

	return msg
}

// NewReader returns a new WAV reader wrapper
func NewReader(rd io.ReadSeeker, size int64) (wav *Reader, err error) {
	if size > maxSize {
		return nil, ErrInputToLarge
	}

	wav = new(Reader)
	wav.input = rd
	wav.size = size

	err = wav.parseHeaders()
	if err != nil {
		return nil, err
	}

	wav.samplesRead = 0

	return wav, nil
}

func (wav *Reader) parseHeaders() (err error) {

	wav.header = &riffHeader{}
	var (
		chunk     [4]byte
		chunkSize uint32
	)

	// decode header
	if err = binary.Read(wav.input, binary.LittleEndian, wav.header); err != nil {
		return err
	}

	if wav.header.Ftype != tokenRiff {
		return ErrNotRiff
	}

	if wav.header.ChunkSize+8 != uint32(wav.size) {
		return ErrIncorrectChunkSize{wav.header.ChunkSize + 8, uint32(wav.size)}
	}

	if wav.header.ChunkFormat != tokenWaveFormat {
		return ErrNotWave
	}

readLoop:
	for {
		// Read next chunkID
		err = binary.Read(wav.input, binary.BigEndian, &chunk)
		if err == io.EOF {
			return io.ErrUnexpectedEOF
		} else if err != nil {
			return err
		}

		// and it's size in bytes
		err = binary.Read(wav.input, binary.LittleEndian, &chunkSize)
		if err == io.EOF {
			return io.ErrUnexpectedEOF
		} else if err != nil {
			return err
		}

		switch chunk {
		case tokenChunkFmt:
			// seek 4 bytes back because riffChunkFmt reads the chunkSize again
			if _, err = wav.input.Seek(-4, os.SEEK_CUR); err != nil {
				return err
			}
			wav.canonical = chunkSize == 16 // canonical format if chunklen == 16
			if err = wav.parseChunkFmt(); err != nil {
				return err
			}
		case tokenData:
			size, _ := wav.input.Seek(0, os.SEEK_CUR)
			wav.firstSamplePos = uint32(size)
			wav.dataBlocSize = uint32(chunkSize)
			break readLoop
		default:
			//fmt.Fprintf(os.Stderr, "Skip unused chunk \"%s\" (%d bytes).\n", chunk, chunkSize)
			wav.extraChunk = true
			if _, err = wav.input.Seek(int64(chunkSize), os.SEEK_CUR); err != nil {
				return err
			}
		}
	}

	if wav.chunkFmt == nil {
		return ErrBrokenChunkFmt
	}

	wav.bytesPerSample = uint32(wav.chunkFmt.BitsPerSample / 8)

	if wav.bytesPerSample == 0 {
		return ErrNoBitsPerSample
	}

	wav.numSamples = wav.dataBlocSize / wav.bytesPerSample
	wav.duration = time.Duration(float64(wav.numSamples)/float64(wav.chunkFmt.SampleRate)) * time.Second

	return nil
}

// parseChunkFmt
func (wav *Reader) parseChunkFmt() (err error) {
	wav.chunkFmt = new(riffChunkFmt)

	if err = binary.Read(wav.input, binary.LittleEndian, wav.chunkFmt); err != nil {
		return err
	}

	if wav.canonical == false {
		var extraparams uint32
		// Get extra params size
		if err = binary.Read(wav.input, binary.LittleEndian, &extraparams); err != nil {
			return err
		}
		// Skip them
		if _, err = wav.input.Seek(int64(extraparams), os.SEEK_CUR); err != nil {
			return err
		}
	}

	// Is audio supported ?
	if wav.chunkFmt.AudioFormat != 1 {
		return ErrFormatNotSupported
	}

	return nil
}

// GetSampleCount returns the number of samples
func (wav *Reader) GetSampleCount() uint32 {
	return wav.numSamples
}

// GetAudioFormat returns the audio format. A value of 1 indicates uncompressed PCM.
// Any other value indicates a compressed format
func (wav *Reader) GetAudioFormat() uint16 {
	return wav.chunkFmt.AudioFormat
}

// GetNumChannels returns the number of audio channels
func (wav *Reader) GetNumChannels() uint16 {
	return wav.chunkFmt.NumChannels
}

// GetSampleRate returns the sample rate
func (wav *Reader) GetSampleRate() uint32 {
	return wav.chunkFmt.SampleRate
}

// GetBitsPerSample returns the number of bits per sample
func (wav *Reader) GetBitsPerSample() uint16 {
	return wav.chunkFmt.BitsPerSample
}

// GetBytesPerSec returns the number of bytes per second of audio. ie: byte rate
func (wav *Reader) GetBytesPerSec() uint32 {
	return wav.chunkFmt.BytesPerSec
}

// GetDuration returns the length of audio
func (wav *Reader) GetDuration() time.Duration {
	return wav.duration
}

// GetFile returns File
func (wav Reader) GetFile() File {
	return File{
		SampleRate:      wav.chunkFmt.SampleRate,
		Channels:        wav.chunkFmt.NumChannels,
		SignificantBits: wav.chunkFmt.BitsPerSample,
		BytesPerSecond:  wav.chunkFmt.BytesPerSec,
		AudioFormat:     wav.chunkFmt.AudioFormat,
		NumberOfSamples: wav.numSamples,
		SoundSize:       wav.dataBlocSize,
		Duration:        wav.duration,
		Canonical:       wav.canonical && !wav.extraChunk,
	}
}

// FirstSampleOffset in the WAV stream
func (wav Reader) FirstSampleOffset() uint32 {
	return wav.firstSamplePos
}

// Reset the wavReader
func (wav *Reader) Reset() (err error) {
	_, err = wav.input.Seek(int64(wav.firstSamplePos), os.SEEK_SET)
	if err == nil {
		wav.samplesRead = 0
	}

	return
}

// GetDumbReader gives you a std io.Reader, starting from the first sample. usefull for piping data.
func (wav Reader) GetDumbReader() (r io.Reader, err error) {
	// move reader to the first sample
	_, err = wav.input.Seek(int64(wav.firstSamplePos), os.SEEK_SET)
	if err != nil {
		return nil, err
	}

	return wav.input, nil
}

// ReadRawSample returns the raw []byte slice
func (wav *Reader) ReadRawSample() ([]byte, error) {
	if wav.samplesRead > wav.numSamples {
		return nil, io.EOF
	}

	buf := make([]byte, wav.bytesPerSample)
	n, err := wav.input.Read(buf)
	if err != nil {
		return nil, err
	}

	if n != int(wav.bytesPerSample) {
		return nil, fmt.Errorf("Read %d bytes, should have read %d", n, wav.bytesPerSample)
	}

	wav.samplesRead++

	return buf, nil
}

// ReadSample returns the parsed sample bytes as integers
func (wav *Reader) ReadSample() (n int32, err error) {
	s, err := wav.ReadRawSample()
	if err != nil {
		return 0, err
	}

	switch wav.bytesPerSample {
	case 1:
		n = int32(s[0])
	case 2:
		n = int32(s[0]) + int32(s[1])<<8
	case 3:
		n = int32(s[0]) + int32(s[1])<<8 + int32(s[2])<<16
	case 4:
		n = int32(s[0]) + int32(s[1])<<8 + int32(s[2])<<16 + int32(s[3])<<24
	default:
		n = 0
		err = fmt.Errorf("Unhandled bytesPerSample! b:%d", wav.bytesPerSample)
	}

	return
}

// ReadSampleEvery returns the parsed sample bytes as integers every X samples
func (wav *Reader) ReadSampleEvery(every uint32, average int) (samples []int32, err error) {

	// Reset any other readers
	err = wav.Reset()
	if err != nil {
		return
	}

	var n int32
	var total int
	total = int(wav.numSamples / every)
	for total >= 0 {
		total = total - 1

		n, err = wav.ReadSample()
		if err != nil {
			return
		}

		// lets average the samples for better accuracy
		// if average > 0 {
		// 	var sum = n
		// 	fmt.Println(n)
		// 	for i := 1; i < average; i++ {
		// 		n, err = wav.ReadSample()
		// 		if err != nil {
		// 			return
		// 		}
		// 		fmt.Println(n)
		// 		sum += n
		// 	}
		// 	fmt.Println("Sum:", sum, "/", int32(average), sum/int32(average))
		// 	n = sum / int32(average)
		// }

		// Median seems to reflect better than average
		if average > 0 {
			var sum = make([]int, average)
			sum[0] = int(n)
			for i := 1; i < average; i++ {
				n, err = wav.ReadSample()
				if err != nil {
					return
				}
				sum[i] = int(n)
			}
			sort.Ints(sum)
			// fmt.Println("Sum:", sum, "[", average/2, "] = ", sum[average/2])
			n = int32(sum[average/2])
		}

		samples = append(samples, n)

		_, err = wav.input.Seek(int64(every), os.SEEK_CUR)
		if err != nil {
			return
		}
	}

	return
}

/*
// Sample of WAV
type Sample struct {
	Offset uint32
	Value  int32
	Second uint32
}

// ReadSampleChannelEvery X samples delivered over a channel
func (wav *Reader) ReadSampleChannelEvery(every uint32) (c chan *Sample, e chan error) {

	// Save resources by making a channel to range over
	c = make(chan *Sample, 100)
	e = make(chan error)

	go func() {

		var err error

		// Reset any other readers
		wav.samplesRead = 0

		// Start from the begining
		_, err = wav.input.Seek(int64(wav.firstSamplePos), os.SEEK_SET)
		if err != nil {
			close(c)
			e <- err
			return
		}

		var n int32
		var total int
		var pos uint32
		total = int(wav.numSamples / every)
		for total >= 0 {
			total = total - 1
			pos += every

			// Read a sample
			n, err = wav.ReadSample()
			if err != nil {
				close(c)
				e <- err
				return
			}

			c <- &Sample{
				Offset: pos,
				Second: pos / wav.chunkFmt.SampleRate,
				Value:  n,
			}

			_, err = wav.input.Seek(int64(every), os.SEEK_CUR)
			if err != nil {
				close(c)
				e <- err
				return
			}
		}

		// Reset any other readers
		wav.samplesRead = 0
	}()

	return
}
*/
