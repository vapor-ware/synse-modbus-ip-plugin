package utils

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCastToType_Ok(t *testing.T) {
	var tests = []struct {
		typeName string
		value    []byte
		expected interface{}
		// expectedLength is only used for strings.
		expectedLength int
	}{
		// unsigned 8-bit integer
		{
			typeName: "u8",
			value:    []byte{1},
			expected: uint8(0x01),
		},
		{
			typeName: "uint8",
			value:    []byte{3},
			expected: uint8(0x03),
		},
		{
			typeName: "u8",
			value:    []byte{0},
			expected: uint8(0x0),
		},
		{
			typeName: "uint8",
			value:    []byte{9},
			expected: uint8(0x09),
		},

		// unsigned 16-bit integer
		{
			typeName: "u16",
			value:    []byte{0, 1},
			expected: uint16(0x01),
		},
		{
			typeName: "uint16",
			value:    []byte{3, 2},
			expected: uint16(0x302),
		},
		{
			typeName: "u16",
			value:    []byte{0, 0},
			expected: uint16(0x0),
		},
		{
			typeName: "uint16",
			value:    []byte{9, 9},
			expected: uint16(0x0909),
		},

		// unsigned 32-bit integer
		{
			typeName: "u32",
			value:    []byte{0, 1, 2, 3},
			expected: uint32(0x10203),
		},
		{
			typeName: "uint32",
			value:    []byte{3, 2, 1, 0},
			expected: uint32(0x3020100),
		},
		{
			typeName: "u32",
			value:    []byte{0, 0, 0, 0},
			expected: uint32(0x0),
		},
		{
			typeName: "uint32",
			value:    []byte{9, 9, 9, 9},
			expected: uint32(0x09090909),
		},

		// unsigned 64-bit integer
		{
			typeName: "u64",
			value:    []byte{0, 1, 2, 3, 4, 5, 6, 7},
			expected: uint64(0x1020304050607),
		},
		{
			typeName: "uint64",
			value:    []byte{7, 6, 5, 4, 3, 2, 1, 0},
			expected: uint64(0x706050403020100),
		},
		{
			typeName: "u64",
			value:    []byte{0, 0, 0, 0, 0, 0, 0, 0},
			expected: uint64(0x0),
		},
		{
			typeName: "uint64",
			value:    []byte{9, 9, 9, 9, 9, 9, 9, 9},
			expected: uint64(0x909090909090909),
		},

		// signed 8-bit integer
		{
			typeName: "s8",
			value:    []byte{1},
			expected: int8(0x1),
		},
		{
			typeName: "int8",
			value:    []byte{3},
			expected: int8(0x3),
		},
		{
			typeName: "s8",
			value:    []byte{0},
			expected: int8(0x0),
		},
		{
			typeName: "int8",
			value:    []byte{9},
			expected: int8(0x9),
		},

		// signed 16-bit integer
		{
			typeName: "s16",
			value:    []byte{0, 1},
			expected: int16(0x1),
		},
		{
			typeName: "int16",
			value:    []byte{3, 2},
			expected: int16(0x302),
		},
		{
			typeName: "s16",
			value:    []byte{0, 0},
			expected: int16(0x0),
		},
		{
			typeName: "int16",
			value:    []byte{9, 9},
			expected: int16(0x909),
		},
		// Negative case.
		{
			typeName: "s16",
			value:    []byte{0xff, 0xff},
			expected: int16(-1),
		},

		// signed 32-bit integer
		{
			typeName: "s32",
			value:    []byte{0, 1, 2, 3},
			expected: int32(0x10203),
		},
		{
			typeName: "int32",
			value:    []byte{3, 2, 1, 0},
			expected: int32(0x3020100),
		},
		{
			typeName: "s32",
			value:    []byte{0, 0, 0, 0},
			expected: int32(0x0),
		},
		{
			typeName: "int32",
			value:    []byte{9, 9, 9, 9},
			expected: int32(0x9090909),
		},

		// signed 64-bit integer
		{
			typeName: "s64",
			value:    []byte{0, 1, 2, 3, 4, 5, 6, 7},
			expected: int64(0x1020304050607),
		},
		{
			typeName: "int64",
			value:    []byte{7, 6, 5, 4, 3, 2, 1, 0},
			expected: int64(0x706050403020100),
		},
		{
			typeName: "s64",
			value:    []byte{0, 0, 0, 0, 0, 0, 0, 0},
			expected: int64(0x0),
		},
		{
			typeName: "int64",
			value:    []byte{9, 9, 9, 9, 9, 9, 9, 9},
			expected: int64(0x909090909090909),
		},

		// 32-bit floating point number
		{
			typeName: "f32",
			value:    []byte{0, 1, 2, 3, 4, 5, 6, 7},
			expected: float32(9.2557e-41),
		},
		{
			typeName: "float32",
			value:    []byte{7, 6, 5, 4, 3, 2, 1, 0},
			expected: float32(1.0082514e-34),
		},
		{
			typeName: "f32",
			value:    []byte{0, 0, 0, 0, 0, 0, 0, 0},
			expected: float32(0),
		},
		{
			typeName: "float32",
			value:    []byte{9, 9, 9, 9, 9, 9, 9, 9},
			expected: float32(1.6495023e-33),
		},

		// 64-bit floating point number
		{
			typeName: "f64",
			value:    []byte{0, 1, 2, 3, 4, 5, 6, 7},
			expected: float64(1.40159977307889e-309),
		},
		{
			typeName: "float64",
			value:    []byte{7, 6, 5, 4, 3, 2, 1, 0},
			expected: float64(7.949928895127363e-275),
		},
		{
			typeName: "f64",
			value:    []byte{0, 0, 0, 0, 0, 0, 0, 0},
			expected: float64(0),
		},
		{
			typeName: "float64",
			value:    []byte{9, 9, 9, 9, 9, 9, 9, 9},
			expected: float64(3.882098286554061e-265),
		},

		// boolean
		{
			typeName: "b",
			value:    []byte{},
			expected: false,
		},
		{
			typeName: "b",
			value:    []byte{0},
			expected: false,
		},
		{
			typeName: "bool",
			value:    []byte{1},
			expected: true,
		},
		{
			typeName: "boolean",
			value:    []byte{0xff},
			expected: true,
		},

		// utf-8 string
		{
			typeName:       "t",
			value:          []byte{},
			expected:       "",
			expectedLength: 0,
		},
		{
			typeName:       "string",
			value:          []byte(nil),
			expected:       "",
			expectedLength: 0,
		},
		{
			typeName:       "utf8",
			value:          []byte{0x31, 0x38, 0x30, 0x35, 0x32, 0x32, 0x30, 0x32, 0x31, 0x38, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected:       "1805220218",
			expectedLength: len("1805220218"),
		},
		{
			typeName:       "t4",
			value:          []byte{0x34, 0x2e, 0x30, 0x2e, 0x36, 0x00, 0x00, 0x00},
			expected:       "4.0.6",
			expectedLength: len("4.0.6"),
		},
		{
			typeName:       "t8",
			value:          []byte{0x34, 0x2e, 0x30, 0x2e, 0x36, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected:       "4.0.6",
			expectedLength: len("4.0.6"),
		},
		{
			typeName:       "t10",
			value:          []byte{0x34, 0x2e, 0x30, 0x2e, 0x36, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected:       "4.0.6",
			expectedLength: len("4.0.6"),
		},
		{
			typeName:       "t12",
			value:          []byte{0x34, 0x2e, 0x30, 0x2e, 0x36, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected:       "4.0.6",
			expectedLength: len("4.0.6"),
		},
		{
			typeName:       "t16",
			value:          []byte{0x34, 0x2e, 0x30, 0x2e, 0x36, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected:       "4.0.6",
			expectedLength: len("4.0.6"),
		},
		{
			typeName:       "t20",
			value:          []byte{0x34, 0x2e, 0x30, 0x2e, 0x36, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			expected:       "4.0.6",
			expectedLength: len("4.0.6"),
		},

		// mac address
		{
			typeName:       "macAddress",
			value:          []byte{0xa0, 0x05, 0x92, 0xf3, 0xff, 0x00},
			expected:       "a0:05:92:f3:ff:00",
			expectedLength: len("a0:05:92:f3:ff:00"),
		},

		// mac address wide
		{
			typeName:       "macAddressWide",
			value:          []byte{0x00, 0x58, 0x00, 0x2f, 0x00, 0x42, 0x00, 0x90, 0x00, 0x22, 0x00, 0xac},
			expected:       "58:2f:42:90:22:ac",
			expectedLength: len("58:2f:42:90:22:ac"),
		},
		// ABCD byte order swapped to CDAB, then convert to float 32.
		{
			typeName: "CDABswapf32",
			value:    []byte{0x25, 0x35, 0x42, 0x95},
			expected: float32(74.572670),
		},
		// cpmModelNumber
		{
			typeName: "cpmModelNumber",
			value:    []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x46, 0x34, 0x30, 0x30, 0x39, 0x30, 0x33, 0x36},
			expected: "F4009036",
		},
		// cpmSerialNumber
		{
			typeName: "cpmSerialNumber",
			value:    []byte{0x2d, 0x53, 0x31, 0x2d, 0x32, 0x31, 0x2d, 0x00, 0x30, 0x30, 0x33, 0x30, 0x31, 0x37, 0x37, 0x00},
			expected: "-S1-21- 0030177",
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%s-%d", tt.typeName, i), func(t *testing.T) {
			actual, err := CastToType(tt.typeName, tt.value)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
			if tt.expectedLength != 0 {
				assert.Equal(t, tt.expectedLength, len(actual.(string)))
			}
		})
	}
}

func TestCastToType_Error(t *testing.T) {
	var tests = []struct {
		typeName string
		value    []byte
	}{
		// Unsupported type name
		{
			typeName: "foo",
			value:    []byte{0},
		},
		// Invalid bytes for signed 32-bit integer
		{
			typeName: "s32",
			value:    []byte{},
		},
		{
			typeName: "int32",
			value:    []byte{},
		},
		// Invalid bytes for signed 64-bit integer
		{
			typeName: "s64",
			value:    []byte{},
		},
		{
			typeName: "int64",
			value:    []byte{},
		},
		// Invalid bytes for mac address
		{
			typeName: "macAddress",
			value:    []byte{0x00, 0x01, 0x02, 0x03, 0x04},
		},
		{
			typeName: "macAddress",
			value:    []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06},
		},
		// Invalid bytes for mac address wide
		{
			typeName: "macAddressWide",
			value:    []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a},
		},
		{
			typeName: "macAddressWide",
			value:    []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c},
		},
		// bytes are deprecated
		{
			typeName: "bytes",
			value:    []byte{},
		},

		{
			typeName: "bytes",
			value:    []byte{0xff},
		},

		{
			typeName: "bytes",
			value:    []byte{0x00, 0x01, 0x02, 0x03},
		},

		{
			typeName: "b16",
			value: []byte{0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0d, 0x0e, 0x0f,
				0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1d, 0x1e, 0x1f},
		},
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%s-%d", tt.typeName, i), func(t *testing.T) {
			_, err := CastToType(tt.typeName, tt.value)
			assert.Error(t, err)
		})
	}
}
