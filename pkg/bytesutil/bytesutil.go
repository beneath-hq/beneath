package bytesutil

import (
	"bytes"
	"encoding/binary"
)

// IntToBytes encodes an int to bytes representation
func IntToBytes(x int64) []byte {
	buf := new(bytes.Buffer)
	err := binary.Write(buf, binary.BigEndian, x)
	if err != nil {
		panic(err)
	}
	return buf.Bytes()
}

// BytesToInt decodes an int encoded with IntToBytes
func BytesToInt(b []byte) int64 {
	return int64(binary.BigEndian.Uint64(b))
}
