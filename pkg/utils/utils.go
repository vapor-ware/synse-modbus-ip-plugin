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

// Int16 converts the bytes to a signed int32.
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

// CastToType takes a typeName, which represents a well-known type, and
// a byte slice and will attempt to cast the byte slice to the named type.
func CastToType(typeName string, value []byte) (interface{}, error) {

	switch strings.ToLower(typeName) {
	case "u16", "uint16":
		// unsigned 16-bit integer
		return Bytes(value).Uint16(), nil

	case "u32", "uint32":
		// unsigned 32-bit integer
		return Bytes(value).Uint32(), nil

	case "u64", "uint64":
		// unsigned 64-bit integer
		return Bytes(value).Uint64(), nil

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

	default:
		return nil, fmt.Errorf("unsupported output data type: %s", typeName)
	}
}

// ConvertFahrenheitToCelsius converts a Farenheit reading to Celsius.
func ConvertFahrenheitToCelsius(farenheit float64) (celsius float64) {
	return (farenheit - 32.0) * 5.0 / 9.0
}

// ConvertEnglishToMetric converts a reading in imperial units to metric.
// This is common for the VEM PLC, which is all imperial units.
func ConvertEnglishToMetric(outputType string, reading interface{}) (result interface{}, err error) {

	switch outputType {
	case "temperature":
		r, ok := reading.(float64)
		if !ok {
			return nil, fmt.Errorf("Unable to convert %T, %v to float64", reading, reading)
		}
		return ConvertFahrenheitToCelsius(r), nil
	default:
		return nil, fmt.Errorf("No english to metric conversion for type %v", outputType)
	}
}
