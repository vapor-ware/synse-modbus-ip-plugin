package devices

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
	"github.com/vapor-ware/synse-sdk/sdk"
)

// The given Device's Data field contains unexpected data which cannot be loaded correctly.
func TestLoadModbusDevices_Action_BadConfig(t *testing.T) {
	defer clearDeviceManagers()

	d := &sdk.Device{
		Data: map[string]interface{}{
			"host": "localhost",
			"port": "5050", // this should be an int
		},
	}

	err := LoadModbusDevices.Action(&sdk.Plugin{}, d)
	assert.Error(t, err)
}

// No manager exists for the handler yet.
func TestLoadModbusDevices_Action_NoManagerForHandler(t *testing.T) {
	defer clearDeviceManagers()

	d := &sdk.Device{
		Data:    map[string]interface{}{"host": "localhost", "port": 5050, "address": 10},
		Handler: "test-handler",
	}

	assert.Empty(t, DeviceManagers)

	err := LoadModbusDevices.Action(&sdk.Plugin{}, d)
	assert.NoError(t, err)
	assert.Len(t, DeviceManagers, 1)
	assert.Contains(t, DeviceManagers, "test-handler")
	assert.Len(t, DeviceManagers["test-handler"], 1)

	manager := DeviceManagers["test-handler"][0]
	assert.Len(t, manager.Devices, 1)
	assert.Len(t, manager.Blocks, 0)
	assert.True(t, manager.sorted)
	assert.False(t, manager.parsed)
}

// A manager does not yet exist for the device.
func TestLoadModbusDevices_Action_NoManagerForDevice(t *testing.T) {
	defer clearDeviceManagers()

	d := &sdk.Device{
		Data:    map[string]interface{}{"host": "localhost", "port": 5050, "address": 10},
		Handler: "test-handler",
		Info:    "dev-1",
	}

	// Create an empty entry for a manager. The device should not match
	// this manager, so it should create a new one.
	DeviceManagers["test-handler"] = []*ModbusDeviceManager{{}}
	assert.Len(t, DeviceManagers, 1)
	assert.Len(t, DeviceManagers["test-handler"], 1)

	err := LoadModbusDevices.Action(&sdk.Plugin{}, d)
	assert.NoError(t, err)
	assert.Len(t, DeviceManagers, 1)
	assert.Contains(t, DeviceManagers, "test-handler")
	assert.Len(t, DeviceManagers["test-handler"], 2)

	// The first manager, added at the top of the test.
	manager1 := DeviceManagers["test-handler"][0]
	assert.Len(t, manager1.Devices, 0)
	assert.Len(t, manager1.Blocks, 0)
	assert.False(t, manager1.sorted)
	assert.False(t, manager1.parsed)

	// The second manager, generated via the Action function.
	manager2 := DeviceManagers["test-handler"][1]
	assert.Len(t, manager2.Devices, 1)
	assert.Len(t, manager2.Blocks, 0)
	assert.True(t, manager2.sorted)
	assert.False(t, manager2.parsed)
}

// A manager does exist for the device.
func TestLoadModbusDevices_Action_DeviceHasManager(t *testing.T) {
	defer clearDeviceManagers()

	d := &sdk.Device{
		Data:    map[string]interface{}{"host": "localhost", "port": 5050, "address": 10},
		Handler: "test-handler",
		Info:    "dev-1",
	}

	// Create an entry for a manager. The added device should match this one. The
	// devices which already exist in this manager have addresses greater than and
	// less than the device being added. Since devices get sorted on add, we should
	// expect that the new device falls between the two existing.
	DeviceManagers["test-handler"] = []*ModbusDeviceManager{
		{
			ModbusDeviceData: config.ModbusDeviceData{
				Host: "localhost",
				Port: 5050,
			},
			Devices: []*ModbusDevice{
				{
					Config: &config.ModbusDeviceData{
						Host:    "localhost",
						Port:    5050,
						Address: 5, // Less than the device being added
					},
				},
				{
					Config: &config.ModbusDeviceData{
						Host:    "localhost",
						Port:    5050,
						Address: 20, // Greater than the device being added
					},
				},
			},
		},
	}

	assert.Len(t, DeviceManagers, 1)
	assert.Len(t, DeviceManagers["test-handler"], 1)
	assert.Len(t, DeviceManagers["test-handler"][0].Devices, 2)

	err := LoadModbusDevices.Action(&sdk.Plugin{}, d)
	assert.NoError(t, err)
	assert.Len(t, DeviceManagers, 1)
	assert.Contains(t, DeviceManagers, "test-handler")
	assert.Len(t, DeviceManagers["test-handler"], 1)

	manager := DeviceManagers["test-handler"][0]
	assert.Len(t, manager.Devices, 3)
	assert.Len(t, manager.Blocks, 0)
	assert.True(t, manager.sorted)
	assert.False(t, manager.parsed)

	assert.Equal(t, uint16(5), manager.Devices[0].Config.Address)
	assert.Equal(t, uint16(10), manager.Devices[1].Config.Address)
	assert.Equal(t, uint16(20), manager.Devices[2].Config.Address)
}
