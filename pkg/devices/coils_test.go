package devices

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-modbus-ip-plugin/internal/testutils"
	"github.com/vapor-ware/synse-sdk/sdk"
)

func TestCoilsHandler_Write_NilDevice(t *testing.T) {
	err := CoilsHandler.Write(nil, &sdk.WriteData{})
	assert.Error(t, err)
}

func TestCoilsHandler_Write_NilData(t *testing.T) {
	err := CoilsHandler.Write(&sdk.Device{}, nil)
	assert.Error(t, err)
}

func TestCoilsHandler_Write_BadAddress(t *testing.T) {
	err := CoilsHandler.Write(&sdk.Device{
		Data: map[string]interface{}{
			"address": "not-an-int",
		},
	}, &sdk.WriteData{})
	assert.Error(t, err)
}

func TestWriteCoil_Ok(t *testing.T) {
	cli := testutils.NewFakeModbusClient()
	data := sdk.WriteData{Data: []byte("true")}

	err := writeCoil(cli, 4, &data)
	assert.NoError(t, err)
}

func TestWriteCoil_DataError(t *testing.T) {
	cli := testutils.NewFakeModbusClient()
	data := sdk.WriteData{Data: []byte("unexpected")}

	err := writeCoil(cli, 4, &data)
	assert.Error(t, err)
}

func TestWriteCoil_WriteError(t *testing.T) {
	cli := testutils.NewFakeModbusClient()
	cli.WithError()
	data := sdk.WriteData{Data: []byte("true")}

	err := writeCoil(cli, 4, &data)
	assert.Error(t, err)
}

func TestGetCoilData(t *testing.T) {
	var cases = []struct {
		name string
		data []byte
		out  uint16
	}{
		{
			name: "0x00 from 0",
			data: []byte("0"),
			out:  0,
		},
		{
			name: "0x00 from false",
			data: []byte("false"),
			out:  0,
		},
		{
			name: "0x00 from False",
			data: []byte("False"),
			out:  0,
		},
		{
			name: "0x00 from FALSE",
			data: []byte("FALSE"),
			out:  0,
		},
		{
			name: "0xFF00 from 1",
			data: []byte("1"),
			out:  0xff00,
		},
		{
			name: "0xFF00 from true",
			data: []byte("true"),
			out:  0xff00,
		},
		{
			name: "0xFF00 from True",
			data: []byte("True"),
			out:  0xff00,
		},
		{
			name: "0xFF00 from TRUE",
			data: []byte("TRUE"),
			out:  0xff00,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			addr, err := getCoilData(tt.data)
			assert.NoError(t, err)
			assert.Equal(t, tt.out, addr)
		})
	}
}

func TestGetCoilDataError(t *testing.T) {
	addr, err := getCoilData([]byte("unexpected"))
	assert.Equal(t, uint16(0), addr)
	assert.Error(t, err)
}
