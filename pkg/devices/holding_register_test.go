package devices

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-modbus-ip-plugin/internal/testutils"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
	"github.com/vapor-ware/synse-sdk/sdk"
)

func TestHoldingRegisterHandler_BulkRead_Error(t *testing.T) {
	defer clearDeviceManagers()

	ctxs, err := HoldingRegisterHandler.BulkRead([]*sdk.Device{})
	assert.Error(t, err)
	assert.Nil(t, ctxs)
}

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

// FIXME (etd): need to re-implement. since client is not stored on the manager
//	 anymore, need to find a different way to manually set it to the fake client.
//func TestBulkReadHoldingRegisters(t *testing.T) {
//	cli := testutils.NewFakeModbusClient()
//	cli.WithResponse([]byte{0x01, 0x02, 0x03, 0x04, 0x00, 0x00})
//
//	cfg := config.ModbusConfig{
//		Address:     0,
//		Width:       1,
//		FailOnError: true,
//		Type:        "u16",
//	}
//	managers := []*ModbusDeviceManager{
//		{
//			ModbusConfig: cfg,
//			Devices: []*ModbusDevice{{
//				Device: &sdk.Device{
//					Output: "status",
//				},
//				Config: &cfg,
//			}},
//			Client: cli,
//			parsed: false,
//			sorted: true,
//		},
//	}
//
//	ctxs, err := bulkReadHoldingRegisters(managers)
//	assert.NoError(t, err)
//	assert.Len(t, ctxs, 1)
//	assert.Len(t, ctxs[0].Reading, 1)
//
//	r := ctxs[0].Reading[0]
//	assert.Equal(t, uint16(0x0102), r.Value)
//	assert.Equal(t, output.Status.Unit, r.Unit)
//	assert.Equal(t, output.Status.Type, r.Type)
//	assert.NotEmpty(t, r.Timestamp)
//	assert.Empty(t, r.Context)
//
//	// Verify the correct number of blocks were created.
//	assert.Len(t, managers[0].Blocks, 1)
//	// Verify the block has the correct number of devices.
//	assert.Len(t, managers[0].Blocks[0].Devices, 1)
//	// Verify that the results were trimmed to the block width
//	assert.Equal(t, []byte{0x01, 0x02}, managers[0].Blocks[0].Results)
//}

// FIXME (etd): need to re-implement. since client is not stored on the manager
//	 anymore, need to find a different way to manually set it to the fake client.
//func TestBulkReadHoldingRegisters2(t *testing.T) {
//	cli := testutils.NewFakeModbusClient()
//	cli.WithResponse([]byte{0x01, 0x02, 0x03, 0x04, 0x00, 0x00})
//
//	cfg := config.ModbusConfig{
//		Address:     0,
//		Width:       2,
//		FailOnError: true,
//		Type:        "u32",
//	}
//	managers := []*ModbusDeviceManager{
//		{
//			ModbusConfig: cfg,
//			Devices: []*ModbusDevice{{
//				Device: &sdk.Device{
//					Output: "status",
//				},
//				Config: &cfg,
//			}},
//			Client: cli,
//			parsed: false,
//			sorted: true,
//		},
//	}
//
//	ctxs, err := bulkReadHoldingRegisters(managers)
//	assert.NoError(t, err)
//	assert.Len(t, ctxs, 1)
//	assert.Len(t, ctxs[0].Reading, 1)
//
//	r := ctxs[0].Reading[0]
//	assert.Equal(t, uint32(0x01020304), r.Value)
//	assert.Equal(t, output.Status.Unit, r.Unit)
//	assert.Equal(t, output.Status.Type, r.Type)
//	assert.NotEmpty(t, r.Timestamp)
//	assert.Empty(t, r.Context)
//
//	// Verify the correct number of blocks were created.
//	assert.Len(t, managers[0].Blocks, 1)
//	// Verify the block has the correct number of devices.
//	assert.Len(t, managers[0].Blocks[0].Devices, 1)
//	// Verify that the results were trimmed to the block width
//	assert.Equal(t, []byte{0x01, 0x02, 0x03, 0x04}, managers[0].Blocks[0].Results)
//}

func TestBulkReadHoldingRegisters_ErrorNewClient(t *testing.T) {
	cfg := config.ModbusConfig{
		Address:     0,
		Width:       1,
		FailOnError: true,
		Type:        "u16",
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
			parsed: false,
			sorted: true, // results must be sorted prior to parsing blocks
		},
	}

	ctxs, err := bulkReadHoldingRegisters(managers)
	assert.Error(t, err)
	assert.Nil(t, ctxs)
}

func TestBulkReadHoldingRegisters_ErrorParseBlocks(t *testing.T) {
	cfg := config.ModbusConfig{
		Host:        "localhost",
		Port:        9876,
		Address:     0,
		Width:       1,
		FailOnError: true,
		Type:        "u16",
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
			parsed: false,
			sorted: false, // results must be sorted prior to parsing blocks
		},
	}

	ctxs, err := bulkReadHoldingRegisters(managers)
	assert.Error(t, err)
	assert.Equal(t, ErrDevicesNotSorted, err)
	assert.Nil(t, ctxs)
}

// FIXME (etd): need to re-implement. since client is not stored on the manager
//	 anymore, need to find a different way to manually set it to the fake client.
//func TestBulkReadHoldingRegisters_ModbusError_FailOnError(t *testing.T) {
//	cli := testutils.NewFakeModbusClient()
//	cli.WithError()
//
//	cfg := config.ModbusConfig{
//		Address:     0,
//		Width:       1,
//		FailOnError: true,
//		Type:        "u16",
//	}
//	managers := []*ModbusDeviceManager{
//		{
//			ModbusConfig: cfg,
//			Devices: []*ModbusDevice{{
//				Device: &sdk.Device{
//					Output: "status",
//				},
//				Config: &cfg,
//			}},
//			Client: cli,
//			parsed: false,
//			sorted: true,
//		},
//	}
//
//	ctxs, err := bulkReadHoldingRegisters(managers)
//	assert.Error(t, err)
//	assert.Nil(t, ctxs)
//}

// FIXME (etd): need to re-implement. since client is not stored on the manager
//	 anymore, need to find a different way to manually set it to the fake client.
//func TestBulkReadHoldingRegisters_ModbusError_NoFailOnError(t *testing.T) {
//	cli := testutils.NewFakeModbusClient()
//	cli.WithError()
//
//	cfg := config.ModbusConfig{
//		Address:     0,
//		Width:       1,
//		FailOnError: false,
//		Type:        "u16",
//	}
//	managers := []*ModbusDeviceManager{
//		{
//			ModbusConfig: cfg,
//			Devices: []*ModbusDevice{{
//				Device: &sdk.Device{
//					Output: "status",
//				},
//				Config: &cfg,
//			}},
//			Client: cli,
//			parsed: false,
//			sorted: true,
//		},
//	}
//
//	ctxs, err := bulkReadHoldingRegisters(managers)
//	assert.NoError(t, err)
//	assert.Len(t, ctxs, 0)
//}

// Make sure that read and write functions are not implemented, just BulkRead.
func TestReadOnlyHoldingRegister(t *testing.T) {
	assert.Nil(t, ReadOnlyHoldingRegisterHandler.Read)
	assert.NotNil(t, ReadOnlyHoldingRegisterHandler.BulkRead)
	assert.Nil(t, ReadOnlyHoldingRegisterHandler.Write)
}
