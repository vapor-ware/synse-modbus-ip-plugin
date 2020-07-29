package testendpoints

//package devices

import (
	"fmt"
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

// Test a bulk read on coils 1-103 with handler coil.
func TestBulkReadCoils_CoilHandler(t *testing.T) {
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

		// *** TODO: This looks like it probably works with all device.Handler == "coil"
		// *** TODO: The read_only_coil looks like it causes trouble.
		device.Handler = "coil"

		/*
			if i == 3 {
				device.Handler = "coil"
			} else {
				device.Handler = "read_only_coil"
			}
		*/

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
	// TODO: Is this call correct? Two different handlers.
	contexts, err := modbusDevices.CoilsHandler.BulkRead(devices)
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
