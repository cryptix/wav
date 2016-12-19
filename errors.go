package wav

import (
	"errors"
	"fmt"
)

var (
	// ErrInputToLarge error
	ErrInputToLarge = errors.New("Input too large")
	// ErrNotRiff error
	ErrNotRiff = errors.New("Not a RIFF file")
	// ErrNotWave error
	ErrNotWave = errors.New("Not a WAVE file")
	// ErrBrokenChunkFmt error
	ErrBrokenChunkFmt = errors.New("could not decode chunkFmt")
	// ErrNoBitsPerSample error
	ErrNoBitsPerSample = errors.New("could not decode chunkFmt")
	// ErrFormatNotSupported error
	ErrFormatNotSupported = errors.New("Format not supported - Only uncompressed PCM currently")
)

// ErrIncorrectChunkSize struct
type ErrIncorrectChunkSize struct {
	Got, Wanted uint32
}

func (e ErrIncorrectChunkSize) Error() string {
	return fmt.Sprintf("Incorrect ChunkSize. Got[%d] Wanted[%d]", e.Got, e.Wanted)
}
