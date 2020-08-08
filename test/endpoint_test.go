package testendpoints

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

// FUTURE: Test a bulk read with the coils we currently (7/28/2020) read on the VEM PLC.
// FUTURE: Same for holding registers.
// FUTURE: Mix of coil, read only coil, input, holding, read only holding registers.

// Should be one network call.
func TestBulkReadCoils_CoilHandlerOnly(t *testing.T) {

	// Create the device slice.
	var devices []*sdk.Device

	// TODO: Sort out the -1 here.
	for i := 1; i <= int(modbusDevices.MaximumRegisterCount)-1; i++ {
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
			Output:  "switch",
			Handler: "coil",
		}

		devices = append(devices, device)
	} // end for

	// Permute device order to test sort.
	permutedDevices := make([]*sdk.Device, len(devices))
	perm := rand.Perm(len(devices))
	for i, v := range perm {
		permutedDevices[v] = devices[i]
	}

	// Do the bulk read.
	modbusDevices.ResetModbusCallCounter()                           // Zero out the modbus call counter.
	assert.Equal(t, uint64(0), modbusDevices.GetModbusCallCounter()) // Verify.
	contexts, err := modbusDevices.CoilsHandler.BulkRead(permutedDevices)
	assert.NoError(t, err)
	assert.Equal(t, uint64(1), modbusDevices.GetModbusCallCounter()) // One modbus call on the wire for this bulk read.
	assert.Equal(t, len(devices), len(contexts))                     // One context per device.

	// Programmatically verify contexts.
	for i := 0; i < len(contexts); i++ {

		// contexts[i].Device
		assert.Equal(t, devices[i].Info, contexts[i].Device.Info)
		// Handler is the same.
		assert.Equal(t, devices[i].Handler, contexts[i].Device.Handler)
		// Address is the same.
		assert.Equal(t, devices[i].Data["address"], contexts[i].Device.Data["address"])

		// contexts[i].Reading
		// One reading per context.
		assert.Equal(t, 1, len(contexts[i].Reading))
		// Reading[0] value is address % 3 == 0
		expectedValue := (devices[i].Data["address"]).(int)%3 == 0
		assert.Equal(t, expectedValue, contexts[i].Reading[0].Value)
	}
}

// Test a bulk read on holding registers with handler holding_register. No read_only_holding_register.
// Should be one network call.
func TestBulkReadHoldingRegisters_HoldingRegisterHandlerOnly(t *testing.T) {
	// Create the device slice.
	var devices []*sdk.Device

	// TODO: Sort out the -2 here.
	for i := 1; i <= int(modbusDevices.MaximumRegisterCount)-2; i++ {
		device := &sdk.Device{
			Info: fmt.Sprintf("Coil %d", i),
			Data: map[string]interface{}{
				"host":        "localhost",
				"port":        1502,
				"type":        "s16",
				"width":       2,
				"failOnError": false,
				"address":     i,
			},
			Output:  "number",
			Handler: "holding_register",
		}

		devices = append(devices, device)
	} // end for

	// Permute device order to test sort.
	permutedDevices := make([]*sdk.Device, len(devices))
	perm := rand.Perm(len(devices))
	for i, v := range perm {
		permutedDevices[v] = devices[i]
	}

	// Call bulk read.
	modbusDevices.ResetModbusCallCounter() // Zero out the modbus call counter.
	contexts, err := modbusDevices.HoldingRegisterHandler.BulkRead(permutedDevices)

	assert.NoError(t, err)
	assert.Equal(t, uint64(1), modbusDevices.GetModbusCallCounter()) // One modbus call on the wire for this bulk read.
	assert.Equal(t, len(devices), len(contexts))                     // One context per device.

	// Validate
	for i := 0; i < len(contexts); i++ {

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
		// Reading[0] value is address.
		expectedValue := (devices[i].Data["address"]).(int)
		assert.Equal(t, expectedValue, int((contexts[i].Reading[0].Value).(int16)))
	}
}

// Test a bulk read on input registers 1-103 with handler input_register.
// Should be one network call.
func TestBulkReadInputRegisters_InputRegisterHandlerOnly(t *testing.T) {
	// Create the device slice.
	var devices []*sdk.Device

	for i := 1; i <= int(modbusDevices.MaximumRegisterCount); i++ {
		device := &sdk.Device{
			Info: fmt.Sprintf("Coil %d", i),
			Data: map[string]interface{}{
				"host":        "localhost",
				"port":        1502,
				"type":        "s16",
				"width":       2,
				"failOnError": false,
				"address":     i,
			},
			Output: "number",
		}

		device.Handler = "input_register"
		devices = append(devices, device)
	} // end for

	// Permute device order to test sort.
	permutedDevices := make([]*sdk.Device, len(devices))
	perm := rand.Perm(len(devices))
	for i, v := range perm {
		permutedDevices[v] = devices[i]
	}

	// Call bulk read.
	modbusDevices.ResetModbusCallCounter() // Zero out the modbus call counter.
	contexts, err := modbusDevices.InputRegisterHandler.BulkRead(permutedDevices)
	assert.NoError(t, err)

	// Validate
	for i := 0; i < len(contexts); i++ {

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
		// Reading[0] value is address.
		expectedValue := (devices[i].Data["address"]).(int)
		assert.Equal(t, expectedValue, int((contexts[i].Reading[0].Value).(int16)))
	}
}

// Test a bulk read on holding registers with handler holding_register. No read_only_holding_register.
// Read registers 1-1255. The VEM currently goes up to 1011 (2020-08-07)
func TestBulkReadHoldingRegisters_1255(t *testing.T) {

	// Create the device slice.
	var devices []*sdk.Device
	for i := 1; i <= 1255; i++ {
		device := &sdk.Device{
			Info: fmt.Sprintf("Coil %d", i),
			Data: map[string]interface{}{
				"host":        "localhost",
				"port":        1502,
				"type":        "s16",
				"width":       2,
				"failOnError": false,
				"address":     i,
			},
			Output:  "number",
			Handler: "holding_register",
		}

		devices = append(devices, device)
	} // end for

	// Permute device order to test sort.
	permutedDevices := make([]*sdk.Device, len(devices))
	perm := rand.Perm(len(devices))
	for i, v := range perm {
		permutedDevices[v] = devices[i]
	}

	// Do the bulk read.
	modbusDevices.ResetModbusCallCounter()                           // Zero out the modbus call counter.
	assert.Equal(t, uint64(0), modbusDevices.GetModbusCallCounter()) // Complete paranoia.
	contexts, err := modbusDevices.HoldingRegisterHandler.BulkRead(permutedDevices)

	// Verify
	assert.NoError(t, err)
	// 11 modbus calls on the wire for this bulk read.
	assert.Equal(t, uint64(11), modbusDevices.GetModbusCallCounter())

	// One context per device.
	assert.Equal(t, len(devices), len(contexts))

	// Programmatically verify contexts.
	for i := 0; i < len(contexts); i++ {

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
		// Reading[0] value is address because of how the data is setup in the emulator.
		expectedValue := (devices[i].Data["address"]).(int)
		assert.Equal(t, expectedValue, int((contexts[i].Reading[0].Value).(int16)))
	}
}
