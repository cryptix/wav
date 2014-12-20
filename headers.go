package wav

const (
	maxSize = 2 << 31
)

var (
	tokenRiff       = [4]byte{'R', 'I', 'F', 'F'}
	tokenWaveFormat = [4]byte{'W', 'A', 'V', 'E'}
	tokenChunkFmt   = [4]byte{'f', 'm', 't', ' '}
	tokenData       = [4]byte{'d', 'a', 't', 'a'}
)

type File struct {
	SampleRate      uint32
	SignificantBits uint16
	Channels        uint16
}

// 12 byte header
type riffHeader struct {
	Ftype       [4]byte
	ChunkSize   uint32
	ChunkFormat [4]byte
}

// 20
type riffChunkFmt struct {
	LengthOfHeader uint32
	AudioFormat    uint16 // 1 = PCM not compressed
	NumChannels    uint16
	SampleRate     uint32
	BytesPerSec    uint32
	BytesPerBloc   uint16
	BitsPerSample  uint16
}
