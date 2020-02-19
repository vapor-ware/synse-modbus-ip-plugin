package devices

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-modbus-ip-plugin/internal/testutils"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
	"github.com/vapor-ware/synse-sdk/sdk"
)

func TestInputRegisterHandler_BulkRead_Error(t *testing.T) {
	defer clearDeviceManagers()

	ctxs, err := InputRegisterHandler.BulkRead([]*sdk.Device{})
	assert.Error(t, err)
	assert.Nil(t, ctxs)
}

// FIXME (etd): need to re-implement. since client is not stored on the manager
//	 anymore, need to find a different way to manually set it to the fake client.
//func TestBulkReadInputRegisters(t *testing.T) {
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
//	ctxs, err := bulkReadInputRegisters(managers)
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
//func TestBulkReadInputRegisters2(t *testing.T) {
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
//	ctxs, err := bulkReadInputRegisters(managers)
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

func TestBulkReadInputRegisters_ErrorNewClient(t *testing.T) {
	cli := testutils.NewFakeModbusClient()
	cli.WithResponse([]byte{0x01, 0x02, 0x03, 0x04, 0x00, 0x00})

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
			sorted: false, // results must be sorted prior to parsing blocks
		},
	}

	ctxs, err := bulkReadInputRegisters(managers)
	assert.Error(t, err)
	assert.Nil(t, ctxs)
}

func TestBulkReadInputRegisters_ErrorParseBlocks(t *testing.T) {
	cli := testutils.NewFakeModbusClient()
	cli.WithResponse([]byte{0x01, 0x02, 0x03, 0x04, 0x00, 0x00})

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

	ctxs, err := bulkReadInputRegisters(managers)
	assert.Error(t, err)
	assert.Equal(t, ErrDevicesNotSorted, err)
	assert.Nil(t, ctxs)
}

// FIXME (etd): need to re-implement. since client is not stored on the manager
//	 anymore, need to find a different way to manually set it to the fake client.
//func TestBulkReadInputRegisters_ModbusError_FailOnError(t *testing.T) {
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
//	ctxs, err := bulkReadInputRegisters(managers)
//	assert.Error(t, err)
//	assert.Nil(t, ctxs)
//}

// FIXME (etd): need to re-implement. since client is not stored on the manager
//	 anymore, need to find a different way to manually set it to the fake client.
//func TestBulkReadInputRegisters_ModbusError_NoFailOnError(t *testing.T) {
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
//	ctxs, err := bulkReadInputRegisters(managers)
//	assert.NoError(t, err)
//	assert.Len(t, ctxs, 0)
//}
