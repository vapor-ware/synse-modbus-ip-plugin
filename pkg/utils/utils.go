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

// Uint8 converts the byte slice to a uint8.
func (b Bytes) Uint8() uint8 {
	return b[0]
}

// Uint16 converts the byte slice to a uint16.
func (b Bytes) Uint16() uint16 {
	return binary.BigEndian.Uint16(b)
}

// Uint32 converts the byte slice to a uint32.
func (b Bytes) Uint32() uint32 {
	return binary.BigEndian.Uint32(b)
}

// Uint64 converts the byte slice to a uint64.
func (b Bytes) Uint64() uint64 {
	return binary.BigEndian.Uint64(b)
}

// Int8 converts the byte slice into an int8.
func (b Bytes) Int8() int8 {
	return int8(b[0])
}

// Int16 converts the bytes to a signed int16.
func (b Bytes) Int16() (out int16, err error) {
	buf := bytes.NewReader(b)
	err = binary.Read(buf, binary.BigEndian, &out)
	return
}

// Int32 converts the byte slice to an int32.
func (b Bytes) Int32() (out int32, err error) {
	buf := bytes.NewReader(b)
	err = binary.Read(buf, binary.BigEndian, &out)
	return
}

// Int64 converts the byte slice to an int64.
func (b Bytes) Int64() (out int64, err error) {
	buf := bytes.NewReader(b)
	err = binary.Read(buf, binary.BigEndian, &out)
	return
}

// Bool converts the byte slice to a bool.
func (b Bytes) Bool() bool {
	if b == nil || len(b) == 0 {
		return false
	}
	return !(b[0] == 0)
}

// Utf8 converts the byte slice to a UTF-8 string.
func (b Bytes) Utf8() string {
	return string(b[:clen(b)])
}

// CastToType takes a typeName, which represents a well-known type, and
// a byte slice and will attempt to cast the byte slice to the named type.
func CastToType(typeName string, value []byte) (interface{}, error) {

	switch strings.ToLower(typeName) {

	case "u8", "uint8":
		// unsigned 8-bit integer
		return Bytes(value).Uint8(), nil

	case "u16", "uint16":
		// unsigned 16-bit integer
		return Bytes(value).Uint16(), nil

	case "u32", "uint32":
		// unsigned 32-bit integer
		return Bytes(value).Uint32(), nil

	case "u64", "uint64":
		// unsigned 64-bit integer
		return Bytes(value).Uint64(), nil

	case "s8", "int8":
		// signed 8-bit integer
		return Bytes(value).Int8(), nil

	case "s16", "int16":
		// signed 16-bit integer
		return Bytes(value).Int16()

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

	case "b", "bool", "boolean":
		// bool
		return Bytes(value).Bool(), nil

	case "t", "string", "utf8":
		// utf-8 string
		return Bytes(value).Utf8(), nil

	default:
		return nil, fmt.Errorf("unsupported output data type: %s", typeName)
	}
}

// clen returns the index of the first NULL byte in n or len(n) if n contains no NULL byte.
// This is from golang syscall, but it is not exported. BSD license.
// https://golang.org/src/syscall/syscall_unix.go
func clen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}
