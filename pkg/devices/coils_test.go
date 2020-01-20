package devices

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-modbus-ip-plugin/internal/testutils"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/output"
)

func TestCoilsHandler_BulkRead_Error(t *testing.T) {
	defer clearDeviceManagers()

	ctxs, err := CoilsHandler.BulkRead([]*sdk.Device{})
	assert.Error(t, err)
	assert.Nil(t, ctxs)
}

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

func TestBulkReadCoils(t *testing.T) {
	cli := testutils.NewFakeModbusClient()
	cli.WithResponse([]byte{0x01, 0x02, 0x03, 0x04, 0x00, 0x00})

	cfg := config.ModbusConfig{
		Address:     0,
		Width:       1,
		FailOnError: true,
		Type:        "b",
	}
	managers := []*ModbusDeviceManager{
		{
			ModbusConfig: cfg,
			Devices: []*ModbusDevice{{
				Device: &sdk.Device{
					Output: "status",
				},
				Config: &cfg,
			}},
			Client: cli,
			parsed: false,
			sorted: true,
		},
	}

	ctxs, err := bulkReadCoils(managers)
	assert.NoError(t, err)
	assert.Len(t, ctxs, 1)
	assert.Len(t, ctxs[0].Reading, 1)

	r := ctxs[0].Reading[0]
	assert.Equal(t, true, r.Value)
	assert.Equal(t, output.Status.Unit, r.Unit)
	assert.Equal(t, output.Status.Type, r.Type)
	assert.NotEmpty(t, r.Timestamp)
	assert.Empty(t, r.Context)

	// Verify the correct number of blocks were created.
	assert.Len(t, managers[0].Blocks, 1)
	// Verify the block has the correct number of devices.
	assert.Len(t, managers[0].Blocks[0].Devices, 1)
	// Verify that the results were trimmed to the block width
	assert.Equal(t, []byte{0x01, 0x02}, managers[0].Blocks[0].Results)
}

func TestBulkReadCoils_ErrorParseBlocks(t *testing.T) {
	cli := testutils.NewFakeModbusClient()
	cli.WithResponse([]byte{0x01, 0x02, 0x03, 0x04, 0x00, 0x00})

	cfg := config.ModbusConfig{
		Address:     0,
		Width:       1,
		FailOnError: true,
		Type:        "b",
	}
	managers := []*ModbusDeviceManager{
		{
			ModbusConfig: cfg,
			Devices: []*ModbusDevice{{
				Device: &sdk.Device{
					Output: "status",
				},
				Config: &cfg,
			}},
			Client: cli,
			parsed: false,
			sorted: false, // results must be sorted prior to parsing blocks
		},
	}

	ctxs, err := bulkReadCoils(managers)
	assert.Error(t, err)
	assert.Equal(t, ErrDevicesNotSorted, err)
	assert.Nil(t, ctxs)
}

func TestBulkReadCoils_ModbusError_FailOnError(t *testing.T) {
	cli := testutils.NewFakeModbusClient()
	cli.WithError()

	cfg := config.ModbusConfig{
		Address:     0,
		Width:       1,
		FailOnError: true,
		Type:        "b",
	}
	managers := []*ModbusDeviceManager{
		{
			ModbusConfig: cfg,
			Devices: []*ModbusDevice{{
				Device: &sdk.Device{
					Output: "status",
				},
				Config: &cfg,
			}},
			Client: cli,
			parsed: false,
			sorted: true,
		},
	}

	ctxs, err := bulkReadCoils(managers)
	assert.Error(t, err)
	assert.Nil(t, ctxs)
}

func TestBulkReadCoils_ModbusError_NoFailOnError(t *testing.T) {
	cli := testutils.NewFakeModbusClient()
	cli.WithError()

	cfg := config.ModbusConfig{
		Address:     0,
		Width:       1,
		FailOnError: false,
		Type:        "b",
	}
	managers := []*ModbusDeviceManager{
		{
			ModbusConfig: cfg,
			Devices: []*ModbusDevice{{
				Device: &sdk.Device{
					Output: "status",
				},
				Config: &cfg,
			}},
			Client: cli,
			parsed: false,
			sorted: true,
		},
	}

	ctxs, err := bulkReadCoils(managers)
	assert.NoError(t, err)
	assert.Len(t, ctxs, 0)
}