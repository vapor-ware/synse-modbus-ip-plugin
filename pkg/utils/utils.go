package utils

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"strings"
)

// Bytes represents a slice of bytes and provides conversion functions
// for the byte slice.
type Bytes []byte

// FIXME (etd)
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

// CastToType takes a typeName, which represents a well-known type, and
// a byte slice and will attempt to cast the byte slice to the named type.
func CastToType(typeName string, value []byte) (interface{}, error) {
	switch strings.ToLower(typeName) {
	case "u32", "uint32":
		// unsigned 32-bit integer
		return Bytes(value).Uint32(), nil

	case "u64", "uint64":
		// unsigned 64-bit integer
		return Bytes(value).Uint64(), nil

	case "s32", "int32":
		// signed 32-bit integer
		return Bytes(value).Int32()

	case "s64", "int64":
		// signed 64-bit integer
		return Bytes(value).Int64()

	case "f32", "float32":
		// 32-bit floating point number
		return Bytes(value).Float32(), nil

	case "f64", "float64":
		// 64-bit floating point number
		return Bytes(value).Float64(), nil

	default:
		return nil, fmt.Errorf("unsupported output data type: %s", typeName)
	}
}
