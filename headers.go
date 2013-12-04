package wav

const (
	WAVmaxSize = 2 << 31
)

var (
	WAVriffType    = [4]byte{'R', 'I', 'F', 'F'}
	WAVchunkFormat = [4]byte{'W', 'A', 'V', 'E'}
	WAVTokenFmt    = "fmt "
	WAVTokenData   = "data"
)

type riffHeader struct {
	Ftype       [4]byte
	ChunkSize   uint32
	ChunkFormat [4]byte
}

type riffChunkFmt struct {
	AudioFormat   uint16 // 1 = PCM not compressed
	NumChannels   uint16
	SampleFreq    uint32
	BytesPerSec   uint32
	BytesPerBloc  uint16
	BitsPerSample uint16
}
