package protocol

import (
	"encoding/binary"
	"math"
)

// Float32FromBytes converts a list of bytes (of length 4) to a float32.
func Float32FromBytes(bytes []byte) float32 {
	bits := binary.BigEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}
