package protocol

import (
	"encoding/binary"
	"math"
	"bytes"
)

type Bytes []byte

// for example, see: https://golang.org/pkg/encoding/binary/#Read
// that might be a better/more consistent way of doing this.


// Float32 converts the byte slice to a float32.
func (b Bytes) Float32() float32 {
	return math.Float32frombits(binary.BigEndian.Uint32(b))
}

// Float64 converts the byte slice to a float64.
func (b Bytes) Float64() float64 {
	return math.Float64frombits(binary.BigEndian.Uint64(b))
}

// Uint32 converts the byte slice to a uint32.
func (b Bytes) Uint32() uint32 {
	return binary.BigEndian.Uint32(b)
}

// Uint64 converts the byte slice to a uint64.
func (b Bytes) Uint64() uint64 {
	return binary.BigEndian.Uint64(b)
}

// Int32 converts the byte slice to an int32.
func (b Bytes) Int32() (out int32, err error) {
	buf := bytes.NewReader(b)
	err = binary.Read(buf, binary.BigEndian, &out)
	if err != nil {
		return
	}
	return
}

// Int64 converts the byte slice to an int64.
func (b Bytes) Int64() (out int64, err error) {
	buf := bytes.NewReader(b)
	err = binary.Read(buf, binary.BigEndian, &out)
	if err != nil {
		return
	}
	return
}

// Float32FromBytes converts a list of bytes (of length 4) to a float32.
func Float32FromBytes(bytes []byte) float32 {
	bits := binary.BigEndian.Uint32(bytes)
	float := math.Float32frombits(bits)
	return float
}
//
//// Float64FromBytes converts a list of bytes (of length 8) to a float64.
//func Float64FromBytes(bytes []byte) float64 {
//	bits := binary.BigEndian.Uint64(bytes)
//	float := math.Float64frombits(bits)
//	return float
//}
//
//func Uint32FromBytes(bytes []byte) uint32 {
//
//}
//
//func Uint64