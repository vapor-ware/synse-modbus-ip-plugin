package devices

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-modbus-ip-plugin/internal/testutils"
	"github.com/vapor-ware/synse-sdk/sdk"
)

func TestHoldingRegisterHandler_Write_NilDevice(t *testing.T) {
	err := HoldingRegisterHandler.Write(nil, &sdk.WriteData{})
	assert.Error(t, err)
}

func TestHoldingRegisterHandler_Write_NilData(t *testing.T) {
	err := HoldingRegisterHandler.Write(&sdk.Device{}, nil)
	assert.Error(t, err)
}

func TestHoldingRegisterHandler_Write_BadAddress(t *testing.T) {
	err := HoldingRegisterHandler.Write(&sdk.Device{
		Data: map[string]interface{}{
			"address": "not-an-int",
		},
	}, &sdk.WriteData{})
	assert.Error(t, err)
}

func TestWriteHoldingRegister_Ok(t *testing.T) {
	cli := testutils.NewFakeModbusClient()
	data := sdk.WriteData{Data: []byte("00")}

	err := writeHoldingRegister(cli, 4, &data)
	assert.NoError(t, err)
}

func TestWriteHoldingRegister_DataError(t *testing.T) {
	cli := testutils.NewFakeModbusClient()
	data := sdk.WriteData{Data: []byte("unexpected")}

	err := writeHoldingRegister(cli, 4, &data)
	assert.Error(t, err)
}

func TestWriteHoldingRegister_WriteError(t *testing.T) {
	cli := testutils.NewFakeModbusClient()
	cli.WithError()
	data := sdk.WriteData{Data: []byte("01")}

	err := writeHoldingRegister(cli, 4, &data)
	assert.Error(t, err)
}

func TestGetHoldingRegisterData(t *testing.T) {
	var cases = []struct {
		name string
		data []byte
		out  uint16
	}{
		{
			name: "valid hex: 0",
			data: []byte("0"),
			out:  0,
		},
		{
			name: "valid hex: 00",
			data: []byte("00"),
			out:  0,
		},
		{
			name: "valid hex: 01",
			data: []byte("01"),
			out:  1,
		},
		{
			name: "valid hex: a",
			data: []byte("a"),
			out:  10,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := getHoldingRegisterData(tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.out, addr)
		})
	}
}

func TestGetHoldingRegisterDataError(t *testing.T) {
	addr, err := getHoldingRegisterData([]byte("unexpected"))
	assert.Equal(t, uint16(0), addr)
	assert.Error(t, err)
}
