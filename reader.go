package wav

import (
	"encoding/binary"
	"fmt"
	"io"
	"math"
	"time"
)

const (
	maxWavSize = 2 << 31
)

type WavReader struct {
	file io.ReadSeeker
	size int64

	header   *riffHeader
	chunkFmt *riffChunkFmt

	// info      *WavInfo
	canonical bool
}

type WavInfo struct {
	data_bloc_size uint32 //
	// Computed values
	extra_chunk      bool          // true if an extra chunk was skipped
	bytes_per_sample uint32        // = bits_per_sample >> 3
	num_samples      uint32        // Total number of samples
	sound_duration   time.Duration //
}

type riffHeader struct {
	ftype       [4]byte
	chunkSize   uint32
	chunkFormat [4]byte
}

type riffChunkFmt struct {
	audioFormat   uint16 // 1 = PCM not compressed
	numChannels   uint16
	sampleFreq    uint32
	bytesPerSec   uint32
	bytesPerBloc  uint16
	bitsPerSample uint16
}

func NewWavReader(rd io.ReadSeeker, size int64) (wav *WavReader, err error) {
	if size > maxWavSize {
		err = fmt.Errorf("Input too large")
		return
	}

	wav = &WavReader{}
	wav.input = rd
	wav.size = size

	err = wav.parseHeaders()

	return
}

func (wav *WavReader) parseHeaders() (err error) {

	wav.header = &riffHeader{}
	var (
		chunk     [4]byte
		chunkSize uint32
	)

	// decode header
	if err = binary.Read(wav.input, binary.LittleEndian, wav.header); err != nil {
		return err
	}

	if string(wav.header.ftype) != "RIFF" {
		return fmt.Errorf("Not a RIFF file")
	}

	if header.chunkSize+8 != uint32(wav.size) {
		return fmt.Errorf("Damaged file. Chunk size != file size.")
	}

	if string(wav.header.chunkFormat) != "WAVE" {
		return fmt.Errorf("Not a WAVE file")
	}

readLoop:
	for {
		// Read next chunkID
		if err = binary.Read(wav.input, binary.BigEndian, &chunk); err != nil {
			return err
		}
		// and it's size in bytes
		if err = binary.Read(wav.input, binary.LittleEndian, &chunkSize); err != nil {
			return err
		}

		switch chunk[:4] {
		case []byte("fmt "):
			wav.canonical = chunkSize == 16 // canonical format if chunklen == 16
			if err = wav.parseChunkFmt(); err != nil {
				return err
			}
		case []byte("data"):
			size, _ := wav.input.Seek(0, os.SEEK_CUR)
			wav.wave_first_sample_pos = uint32(size)
			wav.info.data_bloc_size = uint32(chunkSize)
			break readLoop
		default:
			//fmt.Fprintf(os.Stderr, "Skip unused chunk \"%s\" (%d bytes).\n", chunk, chunkSize)
			wav.info.extra_chunk = true
			if _, err = wav.input.Seek(int64(chunkSize), os.SEEK_CUR); err != nil {
				return err
			}
		}
	}

	// Is audio supported ?
	if wav.chunkFmt.audioFormat != 1 {
		return fmt.Errorf("Only PCM (not compressed) format is supported.")
	}

	// Compute some useful values
	// wav.info.bytes_per_sample = wav.info.bits_per_sample >> 3
	// wav.info.num_samples = wav.info.data_bloc_size / wav.info.bytes_per_sample
	// wav.info.sound_duration = time.Duration(float64(wav.info.data_bloc_size)/float64(wav.info.bytes_per_sec)) * time.Second

	// wav.wave_start_offset = gd.offset
	// wav.wave_start_offset_in_bytes = gd.offset * wav.info.bytes_per_sample

	// wav.samples_for_one_byte = 8 / wav.density

	// payload_samples_space := wav.info.num_samples - wav.wave_start_offset
	// wav.payload_max_size = payload_samples_space / wav.samples_for_one_byte

	return nil
}

// parseChunkFmt
func (wav *WavReader) parseChunkFmt() (err error) {
	wav.chunkFmt = &riffChunkFmt{}

	if err = binary.Read(wav.file, binary.LittleEndian, wav.chunkFmt); err != nil {
		return err
	}

	if wav.canonical == false {
		var extraparams uint32
		// Get extra params size
		if err = binary.Read(wav.file, binary.LittleEndian, &extraparams); err != nil {
			return err
		}
		// Skip them
		if _, err = wav.file.Seek(int64(extraparams), os.SEEK_CUR); err != nil {
			return err
		}
	}

	return nil
}
