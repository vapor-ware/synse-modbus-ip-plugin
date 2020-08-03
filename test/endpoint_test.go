package testendpoints

//package devices

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/goburrow/modbus"
	"github.com/stretchr/testify/assert"
	modbusDevices "github.com/vapor-ware/synse-modbus-ip-plugin/pkg/devices"
	"github.com/vapor-ware/synse-sdk/sdk"
)

var host = "localhost" // run on local machine
var port = 1502

// Make sure the emulator is up.
// Coil data is 0 for addresses where address % 3 == 0.
// Register data is the same as the address.
func TestEmulatorSanity(t *testing.T) {
	connectionString := fmt.Sprintf("%s:%d", host, port)
	client := modbus.TCPClient(connectionString)

	result, err := client.ReadCoils(0, 24) // address, quantity.
	assert.NoError(t, err)
	assert.Equal(t, "[1001001 10010010 100100]", fmt.Sprintf("%0b", result))

	result, err = client.ReadHoldingRegisters(0, 24) // address, quantity.
	assert.NoError(t, err)
	assert.Equal(t, "0000000100020003000400050006000700080009000a000b000c000d000e000f00100011001200130014001500160017", fmt.Sprintf("%x", result))

	result, err = client.ReadInputRegisters(0, 24) // address, quantity.
	assert.NoError(t, err)
	assert.Equal(t, "0000000100020003000400050006000700080009000a000b000c000d000e000f00100011001200130014001500160017", fmt.Sprintf("%x", result))
}

// TODO: Test a bulk read with the coils we currently (7/28/2020) read on the VEM PLC.
// TODO: Same for holding registers.
// TODO: Mix of coil, read only coil, input, holding, read only holding registers.

// TODO: The bug here is that this should be one network round trip for all 103 coils. It's currently 103 round trips.
// Test a bulk read on coils 1-103 with handler coil. No read_only_coil.
func TestBulkReadCoils_CoilHandlerOnly(t *testing.T) {
	// Create the device slice.
	fmt.Printf("Creating devices\n")
	var devices []*sdk.Device

	for i := 1; i <= 103; i++ {
		device := &sdk.Device{
			Info: fmt.Sprintf("Coil %d", i),
			Data: map[string]interface{}{
				"host":        "localhost",
				"port":        1502,
				"type":        "b",
				"width":       1,
				"failOnError": false,
				"address":     i,
			},
			Output: "switch",
			//Type: "switch",
			Handler: "coil",
		}

		// *** TODO: This looks like it probably works with all device.Handler == "coil"
		// *** TODO: The read_only_coil looks like it causes trouble.
		device.Handler = "coil"

		//		if i == 3 {
		//			device.Handler = "coil"
		//		} else {
		//			device.Handler = "read_only_coil"
		//		}

		devices = append(devices, device)
	} // end for

	fmt.Printf("dumping devices:\n")
	for i := 0; i < len(devices); i++ {
		fmt.Printf("device[%d]: %+v\n", i, *(devices[i]))
	}

	/*

		// Load the devices in the thinggy.
		fmt.Printf("Loading devices\n")
		for i := 0; i < len(devices); i++ {
			err := modbusDevices.LoadModbusDevices.Action(&sdk.Plugin{}, devices[i])
			assert.NoError(t, err)
		}
		fmt.Printf("Loaded devices\n")

		fmt.Printf("Dumping DeviceManagers\n")
		//fmt.Printf("DeviceManagers: %T, %+v\n", modbusDevices.DeviceManagers, modbusDevices.DeviceManagers)
		fmt.Printf("DeviceMangers:\n")
		for k, v := range modbusDevices.DeviceManagers {
			fmt.Printf("DeviceManager[%v]:\n", k)
			for i := 0; i < len(v); i++ {
				fmt.Printf("[%d]: %+v\n", i, *v[i])
			}
		}

	*/

	// Permute device order to test sort.
	permutedDevices := make([]*sdk.Device, len(devices))
	perm := rand.Perm(len(devices))
	for i, v := range perm {
		permutedDevices[v] = devices[i]
	}

	fmt.Printf("dumping permuted devices:\n")
	for i := 0; i < len(permutedDevices); i++ {
		fmt.Printf("device[%d]: %+v\n", i, *(permutedDevices[i]))
	}

	fmt.Printf("Calling bulk read\n")
	// TODO: Is this call correct? Two different handlers.
	//contexts, err := modbusDevices.CoilsHandler.BulkRead(devices)
	contexts, err := modbusDevices.CoilsHandler.BulkRead(permutedDevices)
	assert.NoError(t, err)
	assert.Equal(t, len(devices), len(contexts)) // One context per device.

	fmt.Printf("contexts (len %d): %+v\n", len(contexts), contexts)
	fmt.Printf("err: %v\n", err)
	fmt.Printf("Called bulk read\n")

	fmt.Printf("Dumping contexts\n")
	for i := 0; i < len(contexts); i++ {
		fmt.Printf("contexts[%d]: %+v\n", i, contexts[i])
		fmt.Printf("\tDevice: %T, %+v\n", contexts[i].Device, contexts[i].Device)
		fmt.Printf("\tReading: %T, len(%d),  %+v\n", contexts[i].Reading, len(contexts[i].Reading), contexts[i].Reading)

		// Dump readings.
		for j := 0; j < len(contexts[i].Reading); j++ {
			fmt.Printf("\tReading[%d], %T, %+v\n", j, contexts[i].Reading[j], contexts[i].Reading[j])
		}

		// Programmatically verify contexts.
		// contexts[i].Device
		// Context device is the same as in the ordered device list.
		assert.Equal(t, devices[i].Info, contexts[i].Device.Info)
		// Handler is the same.
		assert.Equal(t, devices[i].Handler, contexts[i].Device.Handler)
		// Address is the same.
		assert.Equal(t, devices[i].Data["address"], contexts[i].Device.Data["address"])

		// contexts[i].Reading
		// One reading per context.
		assert.Equal(t, 1, len(contexts[i].Reading))
		// Reading[0] value is address % 3 == 0

		//expectedValue := (devices[i].Data["address'"]).(int) % 3 == 0
		//assert.Equal(t, expectedValue, contexts[i].Reading[0].Value)
		//fmt.Printf("*** address: %T, %+v\n", devices[i].Data["address"], devices[i].Data["address"])
		expectedValue := (devices[i].Data["address"]).(int)%3 == 0
		//fmt.Printf("*** expectedValue: %+v\n", expectedValue)
		//fmt.Printf("*** value: %T, %+v\n", contexts[i].Reading[0].Value, contexts[i].Reading[0].Value)
		assert.Equal(t, expectedValue, contexts[i].Reading[0].Value)
	}
}

/*
// TODO: Bug here is 103 round trips. It should be 1.
// Test a bulk read on holding registers 1-103 with handler holding_register. No read_only_holding_register.
func TestBulkReadHoldingRegisters_HoldingRegisterHandlerOnly(t *testing.T) {
	// Create the device slice.
	fmt.Printf("Creating devices\n")
	var devices []*sdk.Device

	for i := 1; i <= 103; i++ {
		device := &sdk.Device{
			Info: fmt.Sprintf("Coil %d", i),
			Data: map[string]interface{}{
				"host":        "localhost",
				"port":        1502,
				"type":        "b",
				"width":       1,
				"failOnError": false,
				"address":     i,
			},
			Output: "switch",
		}

		device.Handler = "holding_register"
		devices = append(devices, device)
	} // end for

	fmt.Printf("dumping devices:\n")
	for i := 0; i < len(devices); i++ {
		fmt.Printf("device[%d]: %+v\n", i, *(devices[i]))
	}

	// Load the devices in the thinggy.
	fmt.Printf("Loading devices\n")
	for i := 0; i < len(devices); i++ {
		err := modbusDevices.LoadModbusDevices.Action(&sdk.Plugin{}, devices[i])
		assert.NoError(t, err)
	}
	fmt.Printf("Loaded devices\n")

	fmt.Printf("Dumping DeviceManagers\n")
	fmt.Printf("DeviceManagers: %T, %+v\n", modbusDevices.DeviceManagers, modbusDevices.DeviceManagers)

	fmt.Printf("Calling bulk read\n")
	contexts, err := modbusDevices.HoldingRegisterHandler.BulkRead(devices)
	fmt.Printf("contexts (len %d): %+v\n", len(contexts), contexts)
	fmt.Printf("err: %v\n", err)
	fmt.Printf("Called bulk read\n")

	fmt.Printf("Dumping contexts\n")
	assert.NoError(t, err)
	for i := 0; i < len(contexts); i++ {
		fmt.Printf("contexts[%d]: %+v\n", i, contexts[i])
		fmt.Printf("\tReading: %T, len(%d),  %+v\n", contexts[i].Reading, len(contexts[i].Reading), contexts[i].Reading)

		// Dump readings.
		for j := 0; j < len(contexts[i].Reading); j++ {
			fmt.Printf("\tReading[%d], %T, %+v\n", j, contexts[i].Reading[j], contexts[i].Reading[j])
		}
	}
}

// TODO: Bug here is 103 round trips. It should be 1.
// Test a bulk read on input registers 1-103 with handler input_register..
func TestBulkReadInputRegisters_InputRegisterHandlerOnly(t *testing.T) {
	// Create the device slice.
	fmt.Printf("Creating devices\n")
	var devices []*sdk.Device

	for i := 1; i <= 103; i++ {
		device := &sdk.Device{
			Info: fmt.Sprintf("Coil %d", i),
			Data: map[string]interface{}{
				"host":        "localhost",
				"port":        1502,
				"type":        "b",
				"width":       1,
				"failOnError": false,
				"address":     i,
			},
			Output: "switch",
		}

		device.Handler = "input_register"
		devices = append(devices, device)
	} // end for

	fmt.Printf("dumping devices:\n")
	for i := 0; i < len(devices); i++ {
		fmt.Printf("device[%d]: %+v\n", i, *(devices[i]))
	}

	// Load the devices in the thinggy.
	fmt.Printf("Loading devices\n")
	for i := 0; i < len(devices); i++ {
		err := modbusDevices.LoadModbusDevices.Action(&sdk.Plugin{}, devices[i])
		assert.NoError(t, err)
	}
	fmt.Printf("Loaded devices\n")

	fmt.Printf("Dumping DeviceManagers\n")
	fmt.Printf("DeviceManagers: %T, %+v\n", modbusDevices.DeviceManagers, modbusDevices.DeviceManagers)

	fmt.Printf("Calling bulk read\n")
	contexts, err := modbusDevices.InputRegisterHandler.BulkRead(devices)
	fmt.Printf("contexts (len %d): %+v\n", len(contexts), contexts)
	fmt.Printf("err: %v\n", err)
	fmt.Printf("Called bulk read\n")

	fmt.Printf("Dumping contexts\n")
	assert.NoError(t, err)
	for i := 0; i < len(contexts); i++ {
		fmt.Printf("contexts[%d]: %+v\n", i, contexts[i])
		fmt.Printf("\tReading: %T, len(%d),  %+v\n", contexts[i].Reading, len(contexts[i].Reading), contexts[i].Reading)

		// Dump readings.
		for j := 0; j < len(contexts[i].Reading); j++ {
			fmt.Printf("\tReading[%d], %T, %+v\n", j, contexts[i].Reading[j], contexts[i].Reading[j])
		}
	}
}
*/
