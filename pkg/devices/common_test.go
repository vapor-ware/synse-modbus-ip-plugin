package devices

import (
	"testing"

	"github.com/vapor-ware/synse-sdk/sdk/output"

	"github.com/stretchr/testify/suite"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
	"github.com/vapor-ware/synse-sdk/sdk"
)

func TestRunSuites(t *testing.T) {
	suite.Run(t, new(NewModbusDeviceTestSuite))
	suite.Run(t, new(NewModbusClientTestSuite))
	suite.Run(t, new(NewModbusClientFromManagerTestSuite))
	suite.Run(t, new(ModbusDeviceManagerTestSuite))
	suite.Run(t, new(ReadBlockTestSuite))
	suite.Run(t, new(UnpackRegisterReadingTestSuite))
	suite.Run(t, new(UnpackCoilReadingTestSuite))
	suite.Run(t, new(UnpackReadingTestSuite))
}

type NewModbusDeviceTestSuite struct {
	suite.Suite
}

func (suite *NewModbusDeviceTestSuite) TestOK() {
	dev := sdk.Device{
		Data: map[string]interface{}{
			"host": "localhost",
			"port": 5050,
		},
	}

	d, err := NewModbusDevice(&dev)
	suite.NoError(err)
	suite.Equal("localhost", d.Config.Host)
	suite.Equal(5050, d.Config.Port)
}

func (suite *NewModbusDeviceTestSuite) TestError() {
	dev := sdk.Device{
		Data: map[string]interface{}{
			"host": "localhost",
			"port": "5050", // should be int
		},
	}

	d, err := NewModbusDevice(&dev)
	suite.Error(err)
	suite.Nil(d)
}

type NewModbusClientTestSuite struct {
	suite.Suite
}

func (suite *NewModbusClientTestSuite) TestOK() {
	dev := sdk.Device{
		Data: map[string]interface{}{
			"host": "localhost",
			"port": 7777,
		},
	}

	client, err := NewModbusClient(&dev)
	suite.NoError(err)
	suite.NotNil(client)
}

func (suite *NewModbusClientTestSuite) TestDecodeError() {
	dev := sdk.Device{
		Data: map[string]interface{}{
			"host": "localhost",
			"port": "not-an-int",
		},
	}

	client, err := NewModbusClient(&dev)
	suite.Error(err)
	suite.Nil(client)
}

func (suite *NewModbusClientTestSuite) TestNewClientError_FailOnError() {
	dev := sdk.Device{
		Data: map[string]interface{}{
			// missing required field: host
			"port":        7777,
			"failOnError": true,
		},
	}

	client, err := NewModbusClient(&dev)
	suite.Error(err)
	suite.Nil(client)
}

func (suite *NewModbusClientTestSuite) TestNewClientError_NoFailOnError() {
	dev := sdk.Device{
		Data: map[string]interface{}{
			// missing required field: host
			"port":        7777,
			"failOnError": false,
		},
	}

	client, err := NewModbusClient(&dev)
	suite.NoError(err)
	suite.Nil(client)
}

type NewModbusClientFromManagerTestSuite struct {
	suite.Suite
}

func (suite *NewModbusClientFromManagerTestSuite) TestOK() {
	manager := ModbusDeviceManager{
		ModbusConfig: config.ModbusConfig{
			Host: "localhost",
			Port: 5050,
		},
	}

	c, err := newModbusClientFromManager(&manager)
	suite.NoError(err)
	suite.NotNil(c)
}

func (suite *NewModbusClientFromManagerTestSuite) TestError_FailOnError() {
	manager := ModbusDeviceManager{
		ModbusConfig: config.ModbusConfig{
			Host:        "", // requires host
			Port:        5050,
			FailOnError: true,
		},
	}

	c, err := newModbusClientFromManager(&manager)
	suite.Error(err)
	suite.Nil(c)
}

func (suite *NewModbusClientFromManagerTestSuite) TestError_NoFailOnError() {
	manager := ModbusDeviceManager{
		ModbusConfig: config.ModbusConfig{
			Host:        "", // requires host
			Port:        5050,
			FailOnError: false,
		},
	}

	c, err := newModbusClientFromManager(&manager)
	suite.NoError(err)
	suite.Nil(c)
}

type ModbusDeviceManagerTestSuite struct {
	suite.Suite
}

func (suite *ModbusDeviceManagerTestSuite) TestNewManager() {
	seed := ModbusDevice{
		Device: &sdk.Device{Info: "test-device-1"},
		Config: &config.ModbusConfig{
			Host: "localhost",
			Port: 5050,
		},
	}

	manager, err := NewModbusDeviceManager(&seed)
	suite.NoError(err)
	suite.Len(manager.Devices, 1)
	suite.Len(manager.Blocks, 0)
	suite.Equal("localhost", manager.Host)
	suite.Equal(5050, manager.Port)
	suite.NotNil(manager.Client)
}

func (suite *ModbusDeviceManagerTestSuite) TestNewManager_NilSeed() {
	manager, err := NewModbusDeviceManager(nil)
	suite.Error(err)
	suite.Nil(manager)
}

func (suite *ModbusDeviceManagerTestSuite) TestNewManager_Error() {
	seed := ModbusDevice{
		Device: &sdk.Device{Info: "test-device-1"},
		Config: &config.ModbusConfig{
			Host:        "", // host is required
			Port:        5050,
			FailOnError: true,
		},
	}

	manager, err := NewModbusDeviceManager(&seed)
	suite.Error(err)
	suite.Nil(manager)
}

func (suite *ModbusDeviceManagerTestSuite) TestMatchesDevice() {
	manager := ModbusDeviceManager{
		ModbusConfig: config.ModbusConfig{
			Host: "localhost",
			Port: 5050,
		},
	}

	dev := ModbusDevice{
		Config: &config.ModbusConfig{
			Host: "localhost",
			Port: 5050,
		},
	}

	suite.True(manager.MatchesDevice(&dev))
}

func (suite *ModbusDeviceManagerTestSuite) TestMatchesDeviceNil() {
	manager := ModbusDeviceManager{
		ModbusConfig: config.ModbusConfig{
			Host: "localhost",
			Port: 5050,
		},
	}

	suite.False(manager.MatchesDevice(nil))
}

func (suite *ModbusDeviceManagerTestSuite) TestDoesNotMatchDeviceHost() {
	manager := ModbusDeviceManager{
		ModbusConfig: config.ModbusConfig{
			Host: "localhost",
			Port: 5050,
		},
	}

	dev := ModbusDevice{
		Config: &config.ModbusConfig{
			Host: "10.1.1.1",
			Port: 5050,
		},
	}

	suite.False(manager.MatchesDevice(&dev))
}

func (suite *ModbusDeviceManagerTestSuite) TestDoesNotMatchDevicePort() {
	manager := ModbusDeviceManager{
		ModbusConfig: config.ModbusConfig{
			Host: "localhost",
			Port: 5050,
		},
	}

	dev := ModbusDevice{
		Config: &config.ModbusConfig{
			Host: "localhost",
			Port: 5051,
		},
	}

	suite.False(manager.MatchesDevice(&dev))
}

func (suite *ModbusDeviceManagerTestSuite) TestDoesNotMatchDeviceTimeout() {
	manager := ModbusDeviceManager{
		ModbusConfig: config.ModbusConfig{
			Host: "localhost",
			Port: 5050,
		},
	}

	dev := ModbusDevice{
		Config: &config.ModbusConfig{
			Host:    "localhost",
			Port:    5050,
			Timeout: "10s",
		},
	}

	suite.False(manager.MatchesDevice(&dev))
}

func (suite *ModbusDeviceManagerTestSuite) TestDoesNotMatchDeviceFailOnErr() {
	manager := ModbusDeviceManager{
		ModbusConfig: config.ModbusConfig{
			Host: "localhost",
			Port: 5050,
		},
	}

	dev := ModbusDevice{
		Config: &config.ModbusConfig{
			Host:        "localhost",
			Port:        5050,
			FailOnError: true,
		},
	}

	suite.False(manager.MatchesDevice(&dev))
}

func (suite *ModbusDeviceManagerTestSuite) TestAddDevice() {
	manager := ModbusDeviceManager{}

	// Set internal flags to ensure they get reset after calling AddDevice
	manager.sorted = true
	manager.parsed = true
	suite.Len(manager.Devices, 0)
	suite.Len(manager.Blocks, 0)

	dev := &ModbusDevice{}
	manager.AddDevice(dev)

	suite.Len(manager.Devices, 1)
	suite.Len(manager.Blocks, 0)
	suite.False(manager.sorted)
	suite.False(manager.parsed)
}

func (suite *ModbusDeviceManagerTestSuite) TestAddDeviceNil() {
	manager := ModbusDeviceManager{}

	// Set internal flags to ensure they do not get reset after calling AddDevice
	manager.sorted = true
	manager.parsed = true
	suite.Len(manager.Devices, 0)
	suite.Len(manager.Blocks, 0)

	manager.AddDevice(nil)

	suite.Len(manager.Devices, 0)
	suite.Len(manager.Blocks, 0)
	suite.True(manager.sorted)
	suite.True(manager.parsed)
}

func (suite *ModbusDeviceManagerTestSuite) TestSort() {
	d1 := &ModbusDevice{Config: &config.ModbusConfig{Host: "a", Port: 1, Address: 2}}
	d2 := &ModbusDevice{Config: &config.ModbusConfig{Host: "a", Port: 1, Address: 1}}
	d3 := &ModbusDevice{Config: &config.ModbusConfig{Host: "b", Port: 1, Address: 2}}

	manager := ModbusDeviceManager{
		Devices: []*ModbusDevice{d1, d2, d3},
	}
	suite.False(manager.sorted)
	suite.False(manager.parsed)

	manager.Sort()
	suite.True(manager.sorted)
	suite.False(manager.parsed)
	suite.Len(manager.Devices, 3)
	suite.Len(manager.Blocks, 0)
	suite.Equal(d2, manager.Devices[0])
	suite.Equal(d1, manager.Devices[1])
	suite.Equal(d3, manager.Devices[2])
}

func (suite *ModbusDeviceManagerTestSuite) TestParseBlocks_AlreadyParsed() {
	d1 := &ModbusDevice{Config: &config.ModbusConfig{Host: "a", Port: 1, Address: 2}}
	manager := ModbusDeviceManager{
		Devices: []*ModbusDevice{d1},
	}
	manager.parsed = true
	suite.Len(manager.Devices, 1)
	suite.Len(manager.Blocks, 0)

	err := manager.ParseBlocks()
	suite.NoError(err)
	suite.True(manager.parsed)
	suite.Len(manager.Devices, 1)
	suite.Len(manager.Blocks, 0)
}

func (suite *ModbusDeviceManagerTestSuite) TestParseBlocks_NotSorted() {
	d1 := &ModbusDevice{Config: &config.ModbusConfig{Host: "a", Port: 1, Address: 2}}
	manager := ModbusDeviceManager{
		Devices: []*ModbusDevice{d1},
	}
	manager.sorted = false
	suite.Len(manager.Devices, 1)
	suite.Len(manager.Blocks, 0)

	err := manager.ParseBlocks()
	suite.Error(err)
	suite.Equal(ErrDevicesNotSorted, err)
	suite.False(manager.sorted)
	suite.Len(manager.Devices, 1)
	suite.Len(manager.Blocks, 0)
}

func (suite *ModbusDeviceManagerTestSuite) TestParseBlocks_SingleBlock() {
	d1 := &ModbusDevice{
		Device: &sdk.Device{Info: "dev-1"},
		Config: &config.ModbusConfig{Host: "a", Port: 1, Address: 2, Width: 2},
	}
	manager := ModbusDeviceManager{
		Devices: []*ModbusDevice{d1},
	}
	manager.parsed = false
	manager.sorted = true
	suite.Len(manager.Devices, 1)
	suite.Len(manager.Blocks, 0)

	err := manager.ParseBlocks()
	suite.NoError(err)

	suite.True(manager.parsed)
	suite.True(manager.sorted)
	suite.Len(manager.Devices, 1)
	suite.Len(manager.Blocks, 1)

	block := manager.Blocks[0]
	suite.Len(block.Devices, 1)
	suite.Equal(d1, block.Devices[0])
	suite.Empty(block.Results)
	suite.Equal(uint16(2), block.StartRegister)
	suite.Equal(uint16(2), block.RegisterCount)

	suite.Equal(int32(0), d1.Device.SortIndex)
}

func (suite *ModbusDeviceManagerTestSuite) TestParseBlocks_SingleBlockMultipleDevices() {
	d1 := &ModbusDevice{
		Device: &sdk.Device{Info: "dev-1"},
		Config: &config.ModbusConfig{Host: "a", Port: 1, Address: 2, Width: 2},
	}
	d2 := &ModbusDevice{
		Device: &sdk.Device{Info: "dev-2"},
		Config: &config.ModbusConfig{Host: "a", Port: 1, Address: 8, Width: 2},
	}
	manager := ModbusDeviceManager{
		Devices: []*ModbusDevice{d1, d2},
	}
	manager.parsed = false
	manager.sorted = true
	suite.Len(manager.Devices, 2)
	suite.Len(manager.Blocks, 0)

	err := manager.ParseBlocks()
	suite.NoError(err)

	suite.True(manager.parsed)
	suite.True(manager.sorted)
	suite.Len(manager.Devices, 2)
	suite.Len(manager.Blocks, 1)

	block := manager.Blocks[0]
	suite.Len(block.Devices, 2)
	suite.Equal(d1, block.Devices[0])
	suite.Equal(d2, block.Devices[1])
	suite.Empty(block.Results)
	suite.Equal(uint16(2), block.StartRegister)
	suite.Equal(uint16(8), block.RegisterCount)

	suite.Equal(int32(0), d1.Device.SortIndex)
	suite.Equal(int32(1), d2.Device.SortIndex)
}

func (suite *ModbusDeviceManagerTestSuite) TestParseBlocks_MultipleBlocks() {
	d1 := &ModbusDevice{
		Device: &sdk.Device{Info: "dev-1"},
		Config: &config.ModbusConfig{Host: "a", Port: 1, Address: 2, Width: 2},
	}
	d2 := &ModbusDevice{
		Device: &sdk.Device{Info: "dev-2"},
		Config: &config.ModbusConfig{Host: "a", Port: 1, Address: 8, Width: 2},
	}
	d3 := &ModbusDevice{
		Device: &sdk.Device{Info: "dev-3"},
		Config: &config.ModbusConfig{Host: "a", Port: 1, Address: 200, Width: 2},
	}
	manager := ModbusDeviceManager{
		Devices: []*ModbusDevice{d1, d2, d3},
	}
	manager.parsed = false
	manager.sorted = true
	suite.Len(manager.Devices, 3)
	suite.Len(manager.Blocks, 0)

	err := manager.ParseBlocks()
	suite.NoError(err)

	suite.True(manager.parsed)
	suite.True(manager.sorted)
	suite.Len(manager.Devices, 3)
	suite.Len(manager.Blocks, 2)

	block := manager.Blocks[0]
	suite.Len(block.Devices, 2)
	suite.Equal(d1, block.Devices[0])
	suite.Equal(d2, block.Devices[1])
	suite.Empty(block.Results)
	suite.Equal(uint16(2), block.StartRegister)
	suite.Equal(uint16(8), block.RegisterCount)

	block = manager.Blocks[1]
	suite.Len(block.Devices, 1)
	suite.Equal(d3, block.Devices[0])
	suite.Empty(block.Results)
	suite.Equal(uint16(200), block.StartRegister)
	suite.Equal(uint16(2), block.RegisterCount)

	suite.Equal(int32(0), d1.Device.SortIndex)
	suite.Equal(int32(1), d2.Device.SortIndex)
	suite.Equal(int32(2), d3.Device.SortIndex)
}

func (suite *ModbusDeviceManagerTestSuite) TestResetClient() {
	manager := ModbusDeviceManager{
		ModbusConfig: config.ModbusConfig{
			Host:        "localhost",
			Port:        6543,
			FailOnError: true,
		},
	}
	suite.Nil(manager.Client)

	err := manager.ResetClient()
	suite.NoError(err)

	// hold a reference to the first client
	firstClient := manager.Client

	err = manager.ResetClient()
	suite.NoError(err)

	// verify the first client is different than the new one
	suite.False(firstClient == manager.Client)
}

type ReadBlockTestSuite struct {
	suite.Suite
}

func (suite *ReadBlockTestSuite) TestNewReadBlock() {
	seed := ModbusDevice{Config: &config.ModbusConfig{
		Address: 12,
		Width:   2,
	}}

	block := NewReadBlock(&seed)
	suite.Empty(block.Results)
	suite.Equal(uint16(2), block.RegisterCount)
	suite.Equal(uint16(12), block.StartRegister)
	suite.Len(block.Devices, 1)
	suite.Equal(&seed, block.Devices[0])
}

func (suite *ReadBlockTestSuite) TestAdd() {
	block := ReadBlock{
		StartRegister: 13,
		RegisterCount: 2,
	}

	block.Add(&ModbusDevice{Config: &config.ModbusConfig{
		Address: 15,
		Width:   2,
	}})

	suite.Len(block.Devices, 1)
	suite.Equal(uint16(13), block.StartRegister)
	suite.Equal(uint16(4), block.RegisterCount)
	suite.Empty(block.Results)
}

func (suite *ReadBlockTestSuite) TestAddNil() {
	block := ReadBlock{
		StartRegister: 13,
		RegisterCount: 2,
	}

	block.Add(nil)

	suite.Len(block.Devices, 0)
	suite.Equal(uint16(13), block.StartRegister)
	suite.Equal(uint16(2), block.RegisterCount)
	suite.Empty(block.Results)
}

type UnpackRegisterReadingTestSuite struct {
	suite.Suite
}

func (suite *UnpackRegisterReadingTestSuite) TestOK() {
	block := &ReadBlock{
		StartRegister: 0,
		Results:       []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08},
	}
	device := &ModbusDevice{
		Config: &config.ModbusConfig{
			Address:     2,
			Width:       2,
			FailOnError: true,
			Type:        "u32",
		},
	}

	r, err := UnpackRegisterReading(&output.Status, block, device)
	suite.NoError(err)
	suite.Equal(uint32(0x05060708), r.Value)
	suite.Equal(output.Status.Type, r.Type)
	suite.Equal(output.Status.Unit, r.Unit)
	suite.NotEmpty(r.Timestamp)
	suite.Empty(r.Context)
}

func (suite *UnpackRegisterReadingTestSuite) TestOK2() {
	block := &ReadBlock{
		StartRegister: 4,
		Results:       []byte{0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d},
	}
	device := &ModbusDevice{
		Config: &config.ModbusConfig{
			Address:     9,
			Width:       1,
			FailOnError: true,
			Type:        "u16",
		},
	}

	r, err := UnpackRegisterReading(&output.Status, block, device)
	suite.NoError(err)
	suite.Equal(uint16(0x0b0c), r.Value)
	suite.Equal(output.Status.Type, r.Type)
	suite.Equal(output.Status.Unit, r.Unit)
	suite.NotEmpty(r.Timestamp)
	suite.Empty(r.Context)
}

func (suite *UnpackRegisterReadingTestSuite) TestOffsetError_FailOnError() {
	block := &ReadBlock{
		StartRegister: 0,
		Results:       []byte{0x01, 0x02, 0x03, 0x04}, // not enough bytes
	}
	device := &ModbusDevice{
		Config: &config.ModbusConfig{
			Address:     2,
			Width:       2,
			FailOnError: true,
			Type:        "u32",
		},
	}

	r, err := UnpackRegisterReading(&output.Status, block, device)
	suite.Error(err)
	suite.Nil(r)
}

func (suite *UnpackRegisterReadingTestSuite) TestOffsetError_NoFailOnError() {
	block := &ReadBlock{
		StartRegister: 0,
		Results:       []byte{0x01, 0x02, 0x03, 0x04}, // not enough bytes
	}
	device := &ModbusDevice{
		Config: &config.ModbusConfig{
			Address:     2,
			Width:       2,
			FailOnError: false,
			Type:        "u32",
		},
	}

	r, err := UnpackRegisterReading(&output.Status, block, device)
	suite.NoError(err)
	suite.Nil(r)
}

type UnpackCoilReadingTestSuite struct {
	suite.Suite
}

func (suite *UnpackCoilReadingTestSuite) TestOK() {
	block := &ReadBlock{
		StartRegister: 0,
		Results:       []byte{0x01, 0x02, 0x03},
	}
	device := &ModbusDevice{
		Config: &config.ModbusConfig{
			Address:     0x11,
			Width:       1,
			FailOnError: true,
			Type:        "b",
		},
	}

	r, err := UnpackCoilReading(&output.Status, block, device)
	suite.NoError(err)

	// This should be the third byte (0x11 / 8 = index 2, results[2] = 0x03)
	// Bit index = 0x11 % 8 = 1
	// Coil state should then be (0x03 & (0x01 << 1)) != 0 ==> (0b11 & 0b10)
	// which in this case is True
	suite.Equal(true, r.Value)
	suite.Equal(output.Status.Type, r.Type)
	suite.Equal(output.Status.Unit, r.Unit)
	suite.NotEmpty(r.Timestamp)
	suite.Empty(r.Context)
}

func (suite *UnpackCoilReadingTestSuite) TestOK2() {
	block := &ReadBlock{
		StartRegister: 0,
		Results:       []byte{0x01, 0x02, 0x03},
	}
	device := &ModbusDevice{
		Config: &config.ModbusConfig{
			Address:     0xa,
			Width:       1,
			FailOnError: true,
			Type:        "b",
		},
	}

	r, err := UnpackCoilReading(&output.Status, block, device)
	suite.NoError(err)

	// This should be the second byte (0xa / 8 = index 1, results[1] = 0x02)
	// Bit index = 0xa % 8 = 2
	// Coil state should then be (0x02 & (0x01 << 2)) != 0 ==> (0b10 & 0b100) != 0
	// which in this case is False
	suite.Equal(false, r.Value)
	suite.Equal(output.Status.Type, r.Type)
	suite.Equal(output.Status.Unit, r.Unit)
	suite.NotEmpty(r.Timestamp)
	suite.Empty(r.Context)
}

func (suite *UnpackCoilReadingTestSuite) TestIndexOutOfBounds_FailOnError() {
	block := &ReadBlock{
		StartRegister: 0,
		Results:       []byte{0x01, 0x02}, // not enough bytes
	}
	device := &ModbusDevice{
		Config: &config.ModbusConfig{
			Address:     0x13,
			Width:       1,
			FailOnError: true,
			Type:        "b",
		},
	}

	r, err := UnpackCoilReading(&output.Status, block, device)
	suite.Error(err)
	suite.Nil(r)
}

func (suite *UnpackCoilReadingTestSuite) TestIndexOutOfBounds_NoFailOnError() {
	block := &ReadBlock{
		StartRegister: 0,
		Results:       []byte{0x01, 0x02}, // not enough bytes
	}
	device := &ModbusDevice{
		Config: &config.ModbusConfig{
			Address:     0x13,
			Width:       1,
			FailOnError: false,
			Type:        "b",
		},
	}

	r, err := UnpackCoilReading(&output.Status, block, device)
	suite.NoError(err)
	suite.Nil(r)
}

type UnpackReadingTestSuite struct {
	suite.Suite
}

func (suite *UnpackReadingTestSuite) TestOK() {
	device := &ModbusDevice{
		Config: &config.ModbusConfig{
			FailOnError: true,
			Type:        "u16",
		},
	}
	data := []byte{0x00, 0x01}

	r, err := UnpackReading(&output.Status, device, data)
	suite.NoError(err)
	suite.Equal(uint16(0x0001), r.Value)
	suite.Equal(output.Status.Type, r.Type)
	suite.Equal(output.Status.Unit, r.Unit)
	suite.NotEmpty(r.Timestamp)
	suite.Empty(r.Context)
}

func (suite *UnpackReadingTestSuite) TestOK2() {
	device := &ModbusDevice{
		Config: &config.ModbusConfig{
			FailOnError: true,
			Type:        "u32",
		},
	}
	data := []byte{0x00, 0x01, 0x02, 0x03}

	r, err := UnpackReading(&output.Status, device, data)
	suite.NoError(err)
	suite.Equal(uint32(0x00010203), r.Value)
	suite.Equal(output.Status.Type, r.Type)
	suite.Equal(output.Status.Unit, r.Unit)
	suite.NotEmpty(r.Timestamp)
	suite.Empty(r.Context)
}

func (suite *UnpackReadingTestSuite) TestCastError_FailOnError() {
	device := &ModbusDevice{
		Config: &config.ModbusConfig{
			FailOnError: true,
			Type:        "unsupported-type",
		},
	}
	data := []byte{0x00, 0x01}

	r, err := UnpackReading(&output.Status, device, data)
	suite.Error(err)
	suite.Nil(r)
}

func (suite *UnpackReadingTestSuite) TestCastError_NoFailOnError() {
	device := &ModbusDevice{
		Config: &config.ModbusConfig{
			FailOnError: false,
			Type:        "unsupported-type",
		},
	}
	data := []byte{0x00, 0x01}

	r, err := UnpackReading(&output.Status, device, data)
	suite.NoError(err)
	suite.Nil(r)
}
