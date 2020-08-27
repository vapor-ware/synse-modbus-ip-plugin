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

// Test a bulk read on coils with handler coil. No read_only_coil.
// Should be one network call.
func TestBulkReadCoils_CoilHandlerOnly(t *testing.T) {

	// Create the device slice.
	var devices []*sdk.Device

	// Non-zero start register is deliberate here. It matters when unpacking the data.
	for i := 1; i <= int(modbusDevices.MaximumRegisterCount); i++ {
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

	assert.Equal(t, int(modbusDevices.MaximumRegisterCount), len(devices))

	// Permute device order to test sort.
	permutedDevices := make([]*sdk.Device, len(devices))
	perm := rand.Perm(len(devices))
	for i, v := range perm {
		permutedDevices[v] = devices[i]
	}

	// Load the devices in the thinggy.
	modbusDevices.PurgeBulkReadManager()
	for i := 0; i < len(permutedDevices); i++ {
		modbusDevices.AddModbusDevice(nil, permutedDevices[i])
	}

	// Do the bulk read.
	modbusDevices.ResetModbusCallCounter()                           // Zero out the modbus call counter.
	assert.Equal(t, uint64(0), modbusDevices.GetModbusCallCounter()) // Verify.

	contexts, err := modbusDevices.CoilsHandler.BulkRead(nil) // Devices parameter is ignored internally, so passed in nil.
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

	// Non-zero start register is deliberate here. It matters when unpacking the data.
	for i := 1; i <= int(modbusDevices.MaximumRegisterCount); i++ {
		device := &sdk.Device{
			Info: fmt.Sprintf("Holding Register %d", i),
			Data: map[string]interface{}{
				"host":        "localhost",
				"port":        1502,
				"type":        "s16",
				"width":       1,
				"failOnError": false,
				"address":     i,
			},
			Output:  "number",
			Handler: "holding_register",
		}

		devices = append(devices, device)
	} // end for

	assert.Equal(t, int(modbusDevices.MaximumRegisterCount), len(devices))

	// Permute device order to test sort.
	permutedDevices := make([]*sdk.Device, len(devices))
	perm := rand.Perm(len(devices))
	for i, v := range perm {
		permutedDevices[v] = devices[i]
	}

	// Load the devices in the thinggy.
	modbusDevices.PurgeBulkReadManager()
	for i := 0; i < len(permutedDevices); i++ {
		modbusDevices.AddModbusDevice(nil, permutedDevices[i])
	}

	// Call bulk read.
	modbusDevices.ResetModbusCallCounter()                              // Zero out the modbus call counter.
	contexts, err := modbusDevices.HoldingRegisterHandler.BulkRead(nil) // Devices parameter is ignored internally, so passed in nil.

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

	// Non-zero start register is deliberate here. It matters when unpacking the data.
	for i := 1; i <= int(modbusDevices.MaximumRegisterCount); i++ {
		device := &sdk.Device{
			Info: fmt.Sprintf("Input Register %d", i),
			Data: map[string]interface{}{
				"host":        "localhost",
				"port":        1502,
				"type":        "s16",
				"width":       1,
				"failOnError": false,
				"address":     i,
			},
			Output: "number",
		}

		device.Handler = "input_register"
		devices = append(devices, device)
	} // end for

	assert.Equal(t, int(modbusDevices.MaximumRegisterCount), len(devices))

	// Permute device order to test sort.
	permutedDevices := make([]*sdk.Device, len(devices))
	perm := rand.Perm(len(devices))
	for i, v := range perm {
		permutedDevices[v] = devices[i]
	}

	// Load the devices in the thinggy.
	modbusDevices.PurgeBulkReadManager()
	for i := 0; i < len(permutedDevices); i++ {
		modbusDevices.AddModbusDevice(nil, permutedDevices[i])
	}

	// Call bulk read.
	modbusDevices.ResetModbusCallCounter()                            // Zero out the modbus call counter.
	contexts, err := modbusDevices.InputRegisterHandler.BulkRead(nil) // Devices parameter is ignored internally, so passed in nil.
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
// Read registers 0-1254. The VEM currently goes up to 1011 (2020-08-07)
func TestBulkReadHoldingRegisters_1255(t *testing.T) {

	// Create the device slice.
	var devices []*sdk.Device
	for i := 0; i < 1255; i++ {
		device := &sdk.Device{
			Info: fmt.Sprintf("Holding Register %d", i),
			Data: map[string]interface{}{
				"host":        "localhost",
				"port":        1502,
				"type":        "s16",
				"width":       1,
				"failOnError": false,
				"address":     i,
			},
			Output:  "number",
			Handler: "holding_register",
		}

		devices = append(devices, device)
	} // end for

	assert.Equal(t, 1255, len(devices))

	// Permute device order to test sort.
	permutedDevices := make([]*sdk.Device, len(devices))
	perm := rand.Perm(len(devices))
	for i, v := range perm {
		permutedDevices[v] = devices[i]
	}

	// Load the devices in the thinggy.
	modbusDevices.PurgeBulkReadManager()
	for i := 0; i < len(permutedDevices); i++ {
		modbusDevices.AddModbusDevice(nil, permutedDevices[i])
	}

	// Do the bulk read.
	modbusDevices.ResetModbusCallCounter()                              // Zero out the modbus call counter.
	assert.Equal(t, uint64(0), modbusDevices.GetModbusCallCounter())    // Complete paranoia.
	contexts, err := modbusDevices.HoldingRegisterHandler.BulkRead(nil) // Devices parameter is ignored so passed in nil.

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

// Test a bulk read on coils with handler read_only_coil. No coil.
// This is a very different case internally than all coil.
// Should be one network call.
func TestBulkReadCoils_ReadOnlyCoilHandlerOnly(t *testing.T) {

	// Create the device slice.
	var devices []*sdk.Device

	// Non-zero start register is deliberate here. It matters when unpacking the data.
	for i := 1; i <= int(modbusDevices.MaximumRegisterCount); i++ {
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
			Handler: "read_only_coil",
		}

		devices = append(devices, device)
	} // end for

	assert.Equal(t, int(modbusDevices.MaximumRegisterCount), len(devices))

	// Permute device order to test sort.
	permutedDevices := make([]*sdk.Device, len(devices))
	perm := rand.Perm(len(devices))
	for i, v := range perm {
		permutedDevices[v] = devices[i]
	}

	// Load the devices in the thinggy.
	modbusDevices.PurgeBulkReadManager()
	for i := 0; i < len(permutedDevices); i++ {
		modbusDevices.AddModbusDevice(nil, permutedDevices[i])
	}

	// Do the bulk read.
	modbusDevices.ResetModbusCallCounter()                           // Zero out the modbus call counter.
	assert.Equal(t, uint64(0), modbusDevices.GetModbusCallCounter()) // Verify.

	contexts, err := modbusDevices.CoilsHandler.BulkRead(nil) // Devices parameter is ignored internally, so passed in nil.
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

// Test a bulk read on holding registers with handler read_only_holding_register. No holding_register.
// This is a very different case internally than all holding_register.
// Should be one network call.
func TestBulkReadHoldingRegisters_ReadOnlyHoldingRegisterHandlerOnly(t *testing.T) {
	// Create the device slice.
	var devices []*sdk.Device

	// Non-zero start register is deliberate here. It matters when unpacking the data.
	for i := 1; i <= int(modbusDevices.MaximumRegisterCount); i++ {
		device := &sdk.Device{
			Info: fmt.Sprintf("Holding Register %d", i),
			Data: map[string]interface{}{
				"host":        "localhost",
				"port":        1502,
				"type":        "s16",
				"width":       1,
				"failOnError": false,
				"address":     i,
			},
			Output:  "number",
			Handler: "read_only_holding_register",
		}

		devices = append(devices, device)
	} // end for

	assert.Equal(t, int(modbusDevices.MaximumRegisterCount), len(devices))

	// Permute device order to test sort.
	permutedDevices := make([]*sdk.Device, len(devices))
	perm := rand.Perm(len(devices))
	for i, v := range perm {
		permutedDevices[v] = devices[i]
	}

	// Load the devices in the thinggy.
	modbusDevices.PurgeBulkReadManager()
	for i := 0; i < len(permutedDevices); i++ {
		modbusDevices.AddModbusDevice(nil, permutedDevices[i])
	}

	// Call bulk read.
	modbusDevices.ResetModbusCallCounter()                              // Zero out the modbus call counter.
	contexts, err := modbusDevices.HoldingRegisterHandler.BulkRead(nil) // Devices parameter is ignored internally, so passed in nil.

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

// Should be one network call.
func TestBulkReadCoils_ReadOnlyAndReadWrite(t *testing.T) {

	// Create the device slice.
	var devices []*sdk.Device

	// read/write
	for i := 0; i < int(modbusDevices.MaximumRegisterCount)/2; i++ {
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

	// read only
	for i := int(modbusDevices.MaximumRegisterCount) / 2; i < int(modbusDevices.MaximumRegisterCount); i++ {
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
			Handler: "read_only_coil",
		}

		devices = append(devices, device)
	} // end for

	assert.Equal(t, int(modbusDevices.MaximumRegisterCount), len(devices))

	// Permute device order to test sort.
	permutedDevices := make([]*sdk.Device, len(devices))
	perm := rand.Perm(len(devices))
	for i, v := range perm {
		permutedDevices[v] = devices[i]
	}

	// Load the devices in the thinggy.
	modbusDevices.PurgeBulkReadManager()
	for i := 0; i < len(permutedDevices); i++ {
		modbusDevices.AddModbusDevice(nil, permutedDevices[i])
	}

	// Do the bulk read.
	modbusDevices.ResetModbusCallCounter()                           // Zero out the modbus call counter.
	assert.Equal(t, uint64(0), modbusDevices.GetModbusCallCounter()) // Verify.

	// The scheduler will call handlers for coil and read_only_coil, so call them manually here.
	contexts, err := modbusDevices.CoilsHandler.BulkRead(nil)           // Devices parameter is ignored internally, so passed in nil.
	contexts2, err2 := modbusDevices.ReadOnlyCoilsHandler.BulkRead(nil) // Devices parameter is ignored internally, so passed in nil.
	assert.NoError(t, err)
	assert.NoError(t, err2)

	assert.Equal(t, uint64(1), modbusDevices.GetModbusCallCounter()) // One modbus call on the wire for this bulk read.
	assert.Equal(t, len(devices), len(contexts))                     // All reads in the first context.
	assert.Equal(t, 0, len(contexts2))                               // No reads in the second context.

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
func TestBulkReadHoldingRegisters_ReadOnlyAndReadWrite(t *testing.T) {
	// Create the device slice.
	var devices []*sdk.Device

	for i := 0; i < int(modbusDevices.MaximumRegisterCount)/2; i++ {
		device := &sdk.Device{
			Info: fmt.Sprintf("Holding Register %d", i),
			Data: map[string]interface{}{
				"host":        "localhost",
				"port":        1502,
				"type":        "s16",
				"width":       1,
				"failOnError": false,
				"address":     i,
			},
			Output:  "number",
			Handler: "holding_register",
		}

		devices = append(devices, device)
	} // end for

	for i := int(modbusDevices.MaximumRegisterCount) / 2; i < int(modbusDevices.MaximumRegisterCount); i++ {
		device := &sdk.Device{
			Info: fmt.Sprintf("Holding Register %d", i),
			Data: map[string]interface{}{
				"host":        "localhost",
				"port":        1502,
				"type":        "s16",
				"width":       1,
				"failOnError": false,
				"address":     i,
			},
			Output:  "number",
			Handler: "read_only_holding_register",
		}

		devices = append(devices, device)
	} // end for

	assert.Equal(t, int(modbusDevices.MaximumRegisterCount), len(devices))

	// Permute device order to test sort.
	permutedDevices := make([]*sdk.Device, len(devices))
	perm := rand.Perm(len(devices))
	for i, v := range perm {
		permutedDevices[v] = devices[i]
	}

	// Load the devices in the thinggy.
	modbusDevices.PurgeBulkReadManager()
	for i := 0; i < len(permutedDevices); i++ {
		modbusDevices.AddModbusDevice(nil, permutedDevices[i])
	}

	// Call bulk read.
	modbusDevices.ResetModbusCallCounter()                           // Zero out the modbus call counter.
	assert.Equal(t, uint64(0), modbusDevices.GetModbusCallCounter()) // Verify.

	// The scheduler will call handlers for coil and read_only_coil, so call them manually here.
	contexts, err := modbusDevices.HoldingRegisterHandler.BulkRead(nil)           // Devices parameter is ignored internally, so passed in nil.
	contexts2, err2 := modbusDevices.ReadOnlyHoldingRegisterHandler.BulkRead(nil) // Devices parameter is ignored internally, so passed in nil.
	assert.NoError(t, err)
	assert.NoError(t, err2)

	assert.Equal(t, uint64(1), modbusDevices.GetModbusCallCounter()) // One modbus call on the wire for this bulk read.
	assert.Equal(t, len(devices), len(contexts))                     // All reads in the first context.
	assert.Equal(t, 0, len(contexts2))                               // No reads in the second context.

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

// This test maps out the holding registers and coils as they were for the VEM-150 as of 8/31/2020.
func TestVEM150DataSetOriginal(t *testing.T) {
	// Create the device slice and one for coils and one for holding registers.
	var devices []*sdk.Device
	var coilDevices []*sdk.Device
	var holdingDevices []*sdk.Device

	// Holding registers.
	for i := 1; i <= 67; i++ {
		device := &sdk.Device{
			Info: fmt.Sprintf("Holding Register %d", i),
			Data: map[string]interface{}{
				"host":        "localhost",
				"port":        1502,
				"type":        "s16",
				"width":       1,
				"failOnError": false,
				"address":     i,
			},
			Output:  "number",
			Handler: "read_only_holding_register",
		}

		if i == 44 {
			device.Handler = "holding_register" // We write to this one.
		}

		holdingDevices = append(holdingDevices, device)
	} // end for

	for i := 1001; i <= 1011; i++ {
		device := &sdk.Device{
			Info: fmt.Sprintf("Holding Register %d", i),
			Data: map[string]interface{}{
				"host":        "localhost",
				"port":        1502,
				"type":        "s16",
				"width":       1,
				"failOnError": false,
				"address":     i,
			},
			Output:  "number",
			Handler: "read_only_holding_register",
		}

		holdingDevices = append(holdingDevices, device)
	} // end for

	// Coils
	// Non-zero start register is deliberate here. It matters when unpacking the data.
	for i := 1; i <= 103; i++ {
		if i >= 23 && i <= 30 {
			continue // skip
		}
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
			Handler: "read_only_coil",
		}

		if i == 3 {
			device.Handler = "coil" // We write to this one.
		}

		coilDevices = append(coilDevices, device)
	} // end for

	// Merge to devices.
	for i := 0; i < len(holdingDevices); i++ {
		devices = append(devices, holdingDevices[i])
	}
	for i := 0; i < len(coilDevices); i++ {
		devices = append(devices, coilDevices[i])
	}

	// Permute device order to test sort.
	permutedDevices := make([]*sdk.Device, len(devices))
	perm := rand.Perm(len(devices))
	for i, v := range perm {
		permutedDevices[v] = devices[i]
	}

	// Load the devices in the thinggy.
	modbusDevices.PurgeBulkReadManager()
	for i := 0; i < len(permutedDevices); i++ {
		modbusDevices.AddModbusDevice(nil, permutedDevices[i])
	}

	// Do the bulk read.
	modbusDevices.ResetModbusCallCounter()                           // Zero out the modbus call counter.
	assert.Equal(t, uint64(0), modbusDevices.GetModbusCallCounter()) // Verify.

	// The scheduler will call handlers for coil and read_only_coil plus holding and read only holding, so call them manually here.
	contextsCoils, err := modbusDevices.CoilsHandler.BulkRead(nil) // Devices parameter is ignored internally, so passed in nil.
	contextsCoilsEmpty, err2 := modbusDevices.ReadOnlyCoilsHandler.BulkRead(nil)
	contextsHolding, err3 := modbusDevices.HoldingRegisterHandler.BulkRead(nil)
	contextsHoldingEmpty, err4 := modbusDevices.ReadOnlyHoldingRegisterHandler.BulkRead(nil)

	assert.NoError(t, err)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.NoError(t, err4)

	assert.Equal(t, uint64(3), modbusDevices.GetModbusCallCounter()) // Three modbus calls on the wire for this bulk read.
	assert.Equal(t, 95, len(contextsCoils))                          // All coil reads in the first context.
	assert.Equal(t, 0, len(contextsCoilsEmpty))                      // No reads in the second context.
	assert.Equal(t, 67+11, len(contextsHolding))                     // All holding register reads in the third context.
	assert.Equal(t, 0, len(contextsHoldingEmpty))                    // No reads in the fourth context.

	// Programmatically verify contextsCoils. (coil/read_only_coil)
	for i := 0; i < len(contextsCoils); i++ {

		// contexts[i].Device
		assert.Equal(t, coilDevices[i].Info, contextsCoils[i].Device.Info)
		// Handler is the same.
		assert.Equal(t, coilDevices[i].Handler, contextsCoils[i].Device.Handler)
		// Address is the same.
		assert.Equal(t, coilDevices[i].Data["address"], contextsCoils[i].Device.Data["address"])

		// contexts[i].Reading
		// One reading per context.
		assert.Equal(t, 1, len(contextsCoils[i].Reading))
		// Reading[0] value is address.
		expectedValue := (coilDevices[i].Data["address"]).(int)%3 == 0
		assert.Equal(t, expectedValue, contextsCoils[i].Reading[0].Value)
	}

	// Verify contextsHolding. (holding_register/read_only_holding_register)
	for i := 0; i < len(contextsHolding); i++ {
		// contexts[i].Device
		// Context device is the same as in the ordered device list.
		assert.Equal(t, holdingDevices[i].Info, contextsHolding[i].Device.Info)
		// Handler is the same.
		assert.Equal(t, holdingDevices[i].Handler, contextsHolding[i].Device.Handler)
		// Address is the same.
		assert.Equal(t, holdingDevices[i].Data["address"], contextsHolding[i].Device.Data["address"])

		// contexts[i].Reading
		// One reading per context.
		assert.Equal(t, 1, len(contextsHolding[i].Reading))
		// Reading[0] value is address.
		expectedValue := (holdingDevices[i].Data["address"]).(int)
		assert.Equal(t, expectedValue, int((contextsHolding[i].Reading[0].Value).(int16)))
	}
}

// This test maps out the holding registers and and input regsters and coils as they seem to be for the VEM-150 as of 9/1/2020.
func TestVEM150DataSetSecondRev(t *testing.T) {

	// Create the device slice and one for coils and one for holding registers and one for input registers.
	var devices []*sdk.Device
	var coilDevices []*sdk.Device
	var holdingDevices []*sdk.Device
	var inputDevices []*sdk.Device

	// Holding registers 1-44.
	for i := 1; i <= 44; i++ {
		device := &sdk.Device{
			Info: fmt.Sprintf("Holding Register %d", i),
			Data: map[string]interface{}{
				"host":        "localhost",
				"port":        1502,
				"type":        "s16",
				"width":       1,
				"failOnError": false,
				"address":     i,
			},
			Output:  "number",
			Handler: "read_only_holding_register",
		}

		if i == 44 {
			device.Handler = "holding_register" // We write to this one. (Pretty sure. Need to double check.)
		}

		holdingDevices = append(holdingDevices, device)
	} // end for

	// Holding registers 52-87.
	for i := 52; i <= 87; i++ {
		device := &sdk.Device{
			Info: fmt.Sprintf("Holding Register %d", i),
			Data: map[string]interface{}{
				"host":        "localhost",
				"port":        1502,
				"type":        "s16",
				"width":       1,
				"failOnError": false,
				"address":     i,
			},
			Output:  "number",
			Handler: "read_only_holding_register",
		}

		holdingDevices = append(holdingDevices, device)
	} // end for

	// Holding registers 120-138.
	for i := 120; i <= 138; i++ {
		device := &sdk.Device{
			Info: fmt.Sprintf("Holding Register %d", i),
			Data: map[string]interface{}{
				"host":        "localhost",
				"port":        1502,
				"type":        "s16",
				"width":       1,
				"failOnError": false,
				"address":     i,
			},
			Output:  "number",
			Handler: "read_only_holding_register",
		}

		holdingDevices = append(holdingDevices, device)
	} // end for

	// Input registers 1-33.
	for i := 1; i <= 33; i++ {
		device := &sdk.Device{
			Info: fmt.Sprintf("Input Register %d", i),
			Data: map[string]interface{}{
				"host":        "localhost",
				"port":        1502,
				"type":        "s16",
				"width":       1,
				"failOnError": false,
				"address":     i,
			},
			Output:  "number",
			Handler: "input_register",
		}

		inputDevices = append(inputDevices, device)
	} // end for

	// Coils 1-130.
	for i := 1; i <= 130; i++ {
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
			Handler: "read_only_coil",
		}

		if i == 120 {
			device.Handler = "coil" // We write to this one.
		}

		coilDevices = append(coilDevices, device)
	} // end for

	// Merge to devices.
	for i := 0; i < len(holdingDevices); i++ {
		devices = append(devices, holdingDevices[i])
	}
	for i := 0; i < len(inputDevices); i++ {
		devices = append(devices, inputDevices[i])
	}
	for i := 0; i < len(coilDevices); i++ {
		devices = append(devices, coilDevices[i])
	}

	// Permute device order to test sort.
	permutedDevices := make([]*sdk.Device, len(devices))
	perm := rand.Perm(len(devices))
	for i, v := range perm {
		permutedDevices[v] = devices[i]
	}

	// Load the devices in the thinggy.
	modbusDevices.PurgeBulkReadManager()
	for i := 0; i < len(permutedDevices); i++ {
		modbusDevices.AddModbusDevice(nil, permutedDevices[i])
	}

	// Do the bulk read.
	modbusDevices.ResetModbusCallCounter()                           // Zero out the modbus call counter.
	assert.Equal(t, uint64(0), modbusDevices.GetModbusCallCounter()) // Verify.

	// The scheduler will call all handlers, so call them manually here.
	contextsCoils, err := modbusDevices.CoilsHandler.BulkRead(nil) // Devices parameter is ignored internally, so passed in nil.
	contextsCoilsEmpty, err2 := modbusDevices.ReadOnlyCoilsHandler.BulkRead(nil)
	contextsHolding, err3 := modbusDevices.HoldingRegisterHandler.BulkRead(nil)
	contextsHoldingEmpty, err4 := modbusDevices.ReadOnlyHoldingRegisterHandler.BulkRead(nil)
	contextsInput, err5 := modbusDevices.InputRegisterHandler.BulkRead(nil)

	assert.NoError(t, err)
	assert.NoError(t, err2)
	assert.NoError(t, err3)
	assert.NoError(t, err4)
	assert.NoError(t, err5)

	assert.Equal(t, uint64(5), modbusDevices.GetModbusCallCounter()) // Five modbus calls on the wire for this bulk read.
	assert.Equal(t, 130, len(contextsCoils))                         // All coil reads in the first context.
	assert.Equal(t, 0, len(contextsCoilsEmpty))                      // No reads in the second context.
	assert.Equal(t, 99, len(contextsHolding))                        // All holding register reads in the third context.
	assert.Equal(t, 0, len(contextsHoldingEmpty))                    // No reads in the fourth context.
	assert.Equal(t, 33, len(contextsInput))                          // All input register reads in the fifth context.

	// Programmatically verify contextsCoils. (coil/read_only_coil)
	for i := 0; i < len(contextsCoils); i++ {

		// contexts[i].Device
		assert.Equal(t, coilDevices[i].Info, contextsCoils[i].Device.Info)
		// Handler is the same.
		assert.Equal(t, coilDevices[i].Handler, contextsCoils[i].Device.Handler)
		// Address is the same.
		assert.Equal(t, coilDevices[i].Data["address"], contextsCoils[i].Device.Data["address"])

		// contexts[i].Reading
		// One reading per context.
		assert.Equal(t, 1, len(contextsCoils[i].Reading))
		// Reading[0] value is address.
		expectedValue := (coilDevices[i].Data["address"]).(int)%3 == 0
		assert.Equal(t, expectedValue, contextsCoils[i].Reading[0].Value)
	}

	// Verify contextsHolding. (holding_register/read_only_holding_register)
	for i := 0; i < len(contextsHolding); i++ {
		// contexts[i].Device
		// Context device is the same as in the ordered device list.
		assert.Equal(t, holdingDevices[i].Info, contextsHolding[i].Device.Info)
		// Handler is the same.
		assert.Equal(t, holdingDevices[i].Handler, contextsHolding[i].Device.Handler)
		// Address is the same.
		assert.Equal(t, holdingDevices[i].Data["address"], contextsHolding[i].Device.Data["address"])

		// contexts[i].Reading
		// One reading per context.
		assert.Equal(t, 1, len(contextsHolding[i].Reading))
		// Reading[0] value is address.
		expectedValue := (holdingDevices[i].Data["address"]).(int)
		assert.Equal(t, expectedValue, int((contextsHolding[i].Reading[0].Value).(int16)))
	}

	// Verify contextsInput. (input_register)
	for i := 0; i < len(contextsInput); i++ {
		// contexts[i].Device
		// Context device is the same as in the ordered device list.
		assert.Equal(t, inputDevices[i].Info, contextsInput[i].Device.Info)
		// Handler is the same.
		assert.Equal(t, inputDevices[i].Handler, contextsInput[i].Device.Handler)
		// Address is the same.
		assert.Equal(t, inputDevices[i].Data["address"], contextsInput[i].Device.Data["address"])

		// contexts[i].Reading
		// One reading per context.
		assert.Equal(t, 1, len(contextsInput[i].Reading))
		// Reading[0] value is address.
		expectedValue := (inputDevices[i].Data["address"]).(int)
		assert.Equal(t, expectedValue, int((contextsInput[i].Reading[0].Value).(int16)))
	}
}
