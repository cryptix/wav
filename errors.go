package wav

import (
	"errors"
	"fmt"
)

var (
	ErrInputToLarge = errors.New("Input too large")
	ErrNotRiff      = errors.New("Not a RIFF file")
	ErrNotWave      = errors.New("Not a WAVE file")

	ErrBrokenChunkFmt  = errors.New("could not decode chunkFmt")
	ErrNoBitsPerSample = errors.New("could not decode chunkFmt")

	ErrFormatNotSupported = errors.New("Format not supported - Only uncompressed PCM currently")
)

type ErrIncorrectChunkSize struct {
	Got, Wanted uint32
}

func (e ErrIncorrectChunkSize) Error() string {
	return fmt.Sprintf("Incorrect ChunkSize. Got[%d] Wanted[%d]", e.Got, e.Wanted)
}
