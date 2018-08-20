package unity

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"

	"github.com/itchio/lzma"
)

// Insert uncompressSize
// props			1
// dictSize 		4
// uncompressedSize 8
func DecompressLZMARaw(data []byte, uncompressedSize int32) ([]byte, error) {
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, uint64(uncompressedSize))
	data = append(data[:5], append(b, data[5:]...)...)
	return ioutil.ReadAll(lzma.NewReader(bytes.NewBuffer(data)))
}
