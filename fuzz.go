// +build gofuzz

package wav

import (
	"bytes"
	"io"
)

func Fuzz(data []byte) int {
	rd, err := NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		if rd != nil {
			panic("rd != nil on error")
		}
		return 0
	}
	for {
		_, err = rd.ReadSample()
		if err != nil {
			if err == io.EOF {
				break
			}
			return 0
		}
	}
	return 1
}
