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
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%s-%d", tt.typeName, i), func(t *testing.T) {
			actual, err := CastToType(tt.typeName, tt.value)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, actual)
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
	}

	for i, tt := range tests {
		t.Run(fmt.Sprintf("%s-%d", tt.typeName, i), func(t *testing.T) {
			_, err := CastToType(tt.typeName, tt.value)
			assert.Error(t, err)
		})
	}
}
