package devices

import (
	"strings"
	"testing"

	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/outputs"
	"github.com/vapor-ware/synse-sdk/sdk"
)

// testData for raw data from modbus.
// Each data point is the offset index so that we can see that we have the correct offsets.
var testData = []uint8{
	0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f,
	0x10, 0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18, 0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f,
	0x20, 0x21, 0x22, 0x23, 0x24, 0x25, 0x26, 0x27, 0x28, 0x29, 0x2a, 0x2b, 0x2c, 0x2d, 0x2e, 0x2f,
	0x30, 0x31, 0x32, 0x33, 0x34, 0x35, 0x36, 0x37, 0x38, 0x39, 0x3a, 0x3b, 0x3c, 0x3d, 0x3e, 0x3f,
	0x40, 0x41, 0x42, 0x43, 0x44, 0x45, 0x46, 0x47, 0x48, 0x49, 0x4a, 0x4b, 0x4c, 0x4d, 0x4e, 0x4f,
	0x50, 0x51, 0x52, 0x53, 0x54, 0x55, 0x56, 0x57, 0x58, 0x59, 0x5a, 0x5b, 0x5c, 0x5d, 0x5e, 0x5f,
	0x60, 0x61, 0x62, 0x63, 0x64, 0x65, 0x66, 0x67, 0x68, 0x69, 0x6a, 0x6b, 0x6c, 0x6d, 0x6e, 0x6f,
	0x70, 0x71, 0x72, 0x73, 0x74, 0x75, 0x76, 0x77, 0x78, 0x79, 0x7a, 0x7b, 0x7c, 0x7d, 0x7e, 0x7f,
	0x80, 0x81, 0x82, 0x83, 0x84, 0x85, 0x86, 0x87, 0x88, 0x89, 0x8a, 0x8b, 0x8c, 0x8d, 0x8e, 0x8f,
	0x90, 0x91, 0x92, 0x93, 0x94, 0x95, 0x96, 0x97, 0x98, 0x99, 0x9a, 0x9b, 0x9c, 0x9d, 0x9e, 0x9f,
	0xa0, 0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8, 0xa9, 0xaa, 0xab, 0xac, 0xad, 0xae, 0xaf,
	0xb0, 0xb1, 0xb2, 0xb3, 0xb4, 0xb5, 0xb6, 0xb7, 0xb8, 0xb9, 0xba, 0xbb, 0xbc, 0xbd, 0xbe, 0xbf,
	0xc0, 0xc1, 0xc2, 0xc3, 0xc4, 0xc5, 0xc6, 0xc7, 0xc8, 0xc9, 0xca, 0xcb, 0xcc, 0xcd, 0xce, 0xcf,
	0xd0, 0xd1, 0xd2, 0xd3, 0xd4, 0xd5, 0xd6, 0xd7, 0xd8, 0xd9, 0xda, 0xdb, 0xdc, 0xdd, 0xde, 0xdf,
	0xe0, 0xe1, 0xe2, 0xe3, 0xe4, 0xe5, 0xe6, 0xe7, 0xe8, 0xe9, 0xea, 0xeb, 0xec, 0xed, 0xee, 0xef,
	0xf0, 0xf1, 0xf2, 0xf3, 0xf4, 0xf5, 0xf6, 0xf7, 0xf8, 0xf9, 0xfa, 0xfb, 0xfc, 0xfd, 0xfe, 0xff,
}

// dumpDevices is a helper to dump out the devices.
func dumpDevices(t *testing.T, devices []*sdk.Device) {

	t.Logf("--- Dumping devices ---")
	t.Logf("Devices: %#v", devices)
	for i := 0; i < len(devices); i++ {
		t.Logf("---")
		t.Logf("Devices[%v]: %#v", i, devices[i])
		t.Logf("---")
		t.Logf("Devices[%v].Outputs[0]: %#v", i, devices[i].Outputs[0])
		t.Logf("---")
	}
	t.Logf("--- Dumping devices end ---")
}

// dumpBulkReadMap is a helper to dump a bulk read mapping so we can see it.
func dumpBulkReadMap(t *testing.T, bulkReadMap map[ModbusBulkReadKey][]*ModbusBulkRead, keyOrder []ModbusBulkReadKey) {

	t.Logf("--- Dumping bulk read map ---")
	for a := 0; a < len(keyOrder); a++ {
		k := keyOrder[a]
		v := bulkReadMap[k]
		t.Logf("bulkReadMap[%#v]: %#v", k, v)
		reads := bulkReadMap[k]
		readsRequired := len(reads)
		t.Logf("readsRequired: %v", readsRequired)
		for i := 0; i < readsRequired; i++ {
			//readsRequired := len(v)
			read := reads[i]
			t.Logf("read[%v]", i)
			t.Logf("startRegister: 0x%04x", read.StartRegister)
			t.Logf("registerCount: 0x%04x", read.RegisterCount)
			t.Logf("endRegister:   0x%04x", read.StartRegister+read.RegisterCount)
			t.Logf("readResults: len: %v,  %x", len(read.ReadResults), read.ReadResults)
			theDevices := bulkReadMap[k][i].Devices
			t.Logf("bulkReadMap[%#v][%v]: %#v", k, i, theDevices)
			for j := 0; j < len(theDevices); j++ {
				t.Logf("\tdevice[%v]: %#v", j, theDevices[j])
			}
		}
	}

	t.Logf("--- Dumping bulk read map end ---")
}

// dumpReadContexts dumps the read contexts and readings to the console.
func dumpReadContexts(t *testing.T, readContexts []*sdk.ReadContext) {

	t.Logf("--- Dumping read contexts ---")
	t.Logf("readContexts: len %v, %#v", len(readContexts), readContexts)
	for i := 0; i < len(readContexts); i++ {
		readContext := readContexts[i]
		t.Logf("readContexts[%v]: %#v", i, readContext)
		for j := 0; j < len(readContext.Reading); j++ {
			reading := readContext.Reading[j]
			t.Logf("\treadings[%v]: %#v", j, reading)
		}
	}

	t.Logf("--- Dumping read contexts end ---")
}

func dumpReadings(t *testing.T, readings []*sdk.Reading) {
	for i := 0; i < len(readings); i++ {
		t.Logf("reading[%v]: %#v", i, readings[i])
		t.Logf("reading.Value: 0x%04x, type %T", readings[i].Value, readings[i].Value)
	}
}

// populateBulkReadMap populates a bulk read map with raw modbus data.
func populateBulkReadMap(t *testing.T, bulkReadMap map[ModbusBulkReadKey][]*ModbusBulkRead, keyOrder []ModbusBulkReadKey) {
	for a := 0; a < len(keyOrder); a++ {
		k := keyOrder[a]
		v := bulkReadMap[k]
		for i := 0; i < len(v); i++ {
			v[i].ReadResults = testData[0 : 2*v[i].RegisterCount] // Two bytes per register.
		}
	}
}

func verifyReadings(t *testing.T, expected []*sdk.Reading, actual []*sdk.Reading) {
	expectedLen := len(expected)
	actualLen := len(actual)
	if expectedLen != actualLen {
		t.Fatalf("expected %v readings, actual %v readings", expectedLen, actualLen)
	}

	for i := 0; i < expectedLen; i++ {
		reading := actual[i]

		// Validate expected versus actual.
		if (*(expected[i])).Type != (*reading).Type {
			t.Fatalf("reading[%v].Type. expected: %v, actual: %v", i, (*(expected[i])).Type, (*(reading)).Type)
		}
		if (*(expected[i])).Info != (*reading).Info {
			t.Fatalf("reading[%v].Info. expected: %v, actual: %v", i, (*(expected[i])).Info, (*(reading)).Info)
		}
		if (*(expected[i])).Unit != (*reading).Unit {
			t.Fatalf("reading[%v].Unit. expected: %v, actual: %v", i, (*(expected[i])).Unit, (*(reading)).Unit)
		}
		if (*(expected[i])).Value != (*reading).Value {
			t.Fatalf("reading[%v].Value. expected: %v type %T, actual: %v type %T",
				i,
				(*(expected[i])).Value,
				(*(expected[i])).Value,
				(*reading).Value,
				(*reading).Value)
		}
	}
}

// verifySingleNilReading verifies that there is one read context with one
// reading that is nil.
func verifySingleNilReading(t *testing.T, readContexts []*sdk.ReadContext) {
	if len(readContexts) != 1 {
		t.Fatalf("Expected 1 read context, got %v", len(readContexts))
	}
	t.Logf("readContexts[0]: %#v", readContexts[0])

	if len(readContexts[0].Reading) != 1 {
		t.Fatalf("Expected 1 reading, got %v", len(readContexts[0].Reading))
	}

	if readContexts[0].Reading[0] != nil {
		t.Fatalf("Expected nil reading, got %#v", readContexts[0].Reading[0])
	}
}

// Test000 was the initial test for getting this working.
func Test000(t *testing.T) {
	t.Logf("Test000 start")

	// Create devices for testing.
	devices := []*sdk.Device{
		&sdk.Device{
			Kind:   "vem-plc.return.air.temperature.setpoint.temperature",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Return Air Temperature Setpoint",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					Info: "Return Air Temperature Setpoint",
					Data: map[string]interface{}{
						"address": 0x24,
						"width":   1,
						"type":    "s16",
					},
				},
			},
		},
		&sdk.Device{
			Kind:   "vem-plc.return.air.temperature",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Return Air Temperature",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					Info: "Return Air Temperature",
					Data: map[string]interface{}{
						"address": 0x0D,
						"width":   1,
						"type":    "s16",
					},
				},
			},
		},

		&sdk.Device{
			Kind:   "vem-plc.cooling.coil.leaving.air.temperature",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Cooling Coil Leaving Air Temperature",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					Info: "Cooling Coil Leaving Air Temperature",
					Data: map[string]interface{}{
						"address": 0x11,
						"width":   1,
						"type":    "s16",
					},
				},
			},
		},
	}

	dumpDevices(t, devices)

	// Sort the devices and test that that works.
	sorted, deviceMap, err := SortDevices(devices, true)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("sorted: %#v", sorted)
	t.Logf("--- device map ---")
	for i := 0; i < len(sorted); i++ {
		t.Logf("deviceMap[%v]: %#v", sorted[i], deviceMap[sorted[i]])
		t.Logf("---")
	}
	t.Logf("--- device map end ---")

	// Check sorted is in order.
	for i := 0; i < (len(sorted) - 1); i++ {
		if sorted[i].Register > sorted[i+1].Register {
			t.Fatalf("Sorted not in sorted order. sorted: %v. Fail at indexes [%v-%v]. values %v, %v",
				sorted, i, i+1, sorted[i], sorted[i+1])
		}
	}

	t.Logf("--- Mapping bulk read ---")
	bulkReadMap, keyOrder, err := MapBulkRead(devices, false, false)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("bulkReadMap %#v", bulkReadMap)
	t.Logf("--- Mapping bulk read end ---")

	dumpBulkReadMap(t, bulkReadMap, keyOrder)

	// Validate Map.
	// There should only be one map entry.
	if len(bulkReadMap) != 1 {
		t.Fatalf("Only one map entry should be present, got %v", len(bulkReadMap))
	}

	expectedKey := ModbusBulkReadKey{
		Host:                 "10.193.4.250",
		Port:                 502,
		Timeout:              "10s",
		FailOnError:          false,
		MaximumRegisterCount: 0x7b,
	}

	reads := bulkReadMap[expectedKey]
	if len(reads) != 1 {
		t.Fatalf("Only one read should be required, got count %v, %#v", len(reads), reads)
	}

	read := reads[0]
	if read.StartRegister != 0x0d {
		t.Fatalf("expected startRegister 0x0d, got 0x%04x", read.StartRegister)
	}

	if read.RegisterCount != 0x18 {
		t.Fatalf("expected registerCount 0x18, got 0x%04x", read.RegisterCount)
	}

	if len(read.Devices) != 3 {
		t.Fatalf("expected 3 devices, got %v", len(read.Devices))
	}

	// Populate the map to simulate readings and dump.
	populateBulkReadMap(t, bulkReadMap, keyOrder)
	dumpBulkReadMap(t, bulkReadMap, keyOrder)

	// Map the read data to the synse read contexts.
	readContexts, err := MapBulkReadData(bulkReadMap, keyOrder)
	if err != nil {
		t.Fatalf("Failed to map bulk read data, error: %v", err.Error())
	}
	dumpReadContexts(t, readContexts)

	// Verify read contexts and each reading.
	if len(readContexts) != 3 {
		t.Fatalf("expected 3 readContexts, got %v", len(readContexts))
	}

	if len(readContexts[0].Reading) != 1 {
		t.Fatalf("expected 1 reading in readContexts[0], got %v", len(readContexts[0].Reading))
	}

	reading := readContexts[0].Reading[0]
	t.Logf("reading: %#v", reading)
	t.Logf("reading.Value: 0x%04x, type %T", reading.Value, reading.Value)
	if reading.Value != int16(0x0001) {
		t.Fatalf("expected reading.Value 0x0001, got 0x%04x, type %T", reading.Value, reading.Value)
	}

	if len(readContexts[1].Reading) != 1 {
		t.Fatalf("expected 1 reading in readContexts[1], got %v", len(readContexts[1].Reading))
	}
	reading = readContexts[1].Reading[0]
	t.Logf("reading: %#v", reading)
	t.Logf("reading.Value: 0x%04x, type %T", reading.Value, reading.Value)
	if reading.Value != int16(0x0809) {
		t.Fatalf("expected reading.Value 0x0809, got 0x%04x", reading.Value)
	}

	if len(readContexts[2].Reading) != 1 {
		t.Fatalf("expected 1 reading in readContexts[2], got %v", len(readContexts[2].Reading))
	}
	reading = readContexts[2].Reading[0]
	t.Logf("reading: %#v", reading)
	t.Logf("reading.Value: 0x%04x, type %T", reading.Value, reading.Value)
	if reading.Value != int16(0x2e2f) {
		t.Fatalf("expected reading.Value 0x2e2f, got 0x%04x", reading.Value)
	}

	t.Logf("Test000 end")
}

// TestVEM tests devices as the modbus over ip configuration on the VEM.
// TODO: Need to add 6 e-gauge devices.
// TODO: Need to add the carousel controller.
func TestVEM(t *testing.T) {
	t.Logf("TestVEM start")

	// Create devices for testing.

	// Holding Registers

	registerDevices := []*sdk.Device{
		&sdk.Device{
			Kind:   "vem-plc.hrc.mixed.fluid.temperature",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "HRC Mixed Fluid Temperature",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Temperature,
					Info:       "HRC Mixed Fluid Temperature",
					Data: map[string]interface{}{
						"address": 0x01,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.loop.entering.fluid.temperature",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Loop Entering Fluid Temperature",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Temperature,
					Info:       "Loop Entering Fluid Temperature",
					Data: map[string]interface{}{
						"address": 0x02,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.valve2.flow",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Minimum Flow Control Valve2 Feedback",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.FlowGpm,
					Info:       "Minimum Flow Control Valve2 Feedback",
					Data: map[string]interface{}{
						"address": 0x05,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.system.fluid.flow",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "System Fluid Flow",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.FlowGpm,
					Info:       "System Fluid Flow",
					Data: map[string]interface{}{
						"address": 0x06,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.server.rack.differential.pressure",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Server Rack Differential Pressure",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.InWCThousanths,
					Info:       "Server Rack Differential Pressure",
					Data: map[string]interface{}{
						"address": 0x07,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.system.leaving.fluid.temperature",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "System Leaving Fluid Temperature",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Temperature,
					Info:       "System Leaving Fluid Temperature",
					Data: map[string]interface{}{
						"address": 0x0A,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.return.air.temperature",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Return Air Temperature",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Temperature,
					Info:       "Return Air Temperature",
					Data: map[string]interface{}{
						"address": 0x0D,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.outdoor.air.temperature",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Outdoor Air Temperature",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Temperature,
					Info:       "Outdoor Air Temperature",
					Data: map[string]interface{}{
						"address": 0x0F,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.cooling.coil.leaving.air.temperature",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Cooling Coil Leaving Air Temperature",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Temperature,
					Info:       "Cooling Coil Leaving Air Temperature",
					Data: map[string]interface{}{
						"address": 0x11,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.dx.discharge.gas.pressure",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "DX Discharge Gas Pressure",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.PsiTenths,
					Info:       "DX Discharge Gas Pressure",
					Data: map[string]interface{}{
						"address": 0x18,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.return.air.temperature.setpoint.temperature",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Return Air Temperature Setpoint",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Temperature,
					Info:       "Return Air Temperature Setpoint",
					Data: map[string]interface{}{
						"address": 0x24,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.hrf.speed.command.fan",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "HRF Speed Command",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.FanSpeedPercent,
					Info:       "HRF Speed Command",
					Data: map[string]interface{}{
						"address": 0x2B,
						"width":   1,
						"type":    "u16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.fan",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "VEM Fan Speed Control",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.FanSpeedPercent,
					Info:       "VEM Fan Speed Control",
					Data: map[string]interface{}{
						"address": 0x2C,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.active.flow",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Active Flow Setpoint",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.FlowGpmTenths,
					Info:       "Active Flow Setpoint",
					Data: map[string]interface{}{
						"address": 0x2D,
						"width":   1,
						"type":    "u16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.fan-speed-actual",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "VEM Fan Speed Actual",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.FanSpeedPercent,
					Info:       "VEM Fan Speed Actual",
					Data: map[string]interface{}{
						"address": 0x32,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.system.flow",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Total System Flow",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.FlowGpmTenths,
					Info:       "Total System Flow",
					Data: map[string]interface{}{
						"address": 0x33,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.fan_minimum",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "VEM Fan Speed Minimum",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.FanSpeedPercentTenths,
					Info:       "VEM Fan Speed Minimum",
					Data: map[string]interface{}{
						"address": 0x40,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},
	}

	// Coils:

	coilDevices := []*sdk.Device{
		&sdk.Device{
			Kind:   "vem-plc.bms.start.switch",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "BMS Start",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Coil,
					Info:       "BMS Start",
					Data: map[string]interface{}{
						"address": 0x03,
						"width":   1,
						"type":    "b",
					},
				},
			},
			Handler: &CoilsHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.compressorA.safety.shutdown.switch",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Compressor Bank A in Safety Shutdown",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Coil,
					Info:       "Compressor Bank A in Safety Shutdown",
					Data: map[string]interface{}{
						"address": 0x21,
						"width":   1,
						"type":    "b",
					},
				},
			},
			Handler: &CoilsHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.compressorB.safety.shutdown.switch",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Compressor Bank B in Safety Shutdown",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Coil,
					Info:       "Compressor Bank B in Safety Shutdown",
					Data: map[string]interface{}{
						"address": 0x22,
						"width":   1,
						"type":    "b",
					},
				},
			},
			Handler: &CoilsHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.system.mode.stage3.switch",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "System Mode Stage3",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Coil,
					Info:       "System Mode Stage3",
					Data: map[string]interface{}{
						"address": 0x25,
						"width":   1,
						"type":    "b",
					},
				},
			},
			Handler: &CoilsHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.system.mode.stage2.switch",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "System Mode Stage2",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Coil,
					Info:       "System Mode Stage2",
					Data: map[string]interface{}{
						"address": 0x26,
						"width":   1,
						"type":    "b",
					},
				},
			},
			Handler: &CoilsHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.keep.alive.switch",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "BMS Keep Alive",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Coil,
					Info:       "BMS Keep Alive",
					Data: map[string]interface{}{
						"address": 0x27,
						"width":   1,
						"type":    "b",
					},
				},
			},
			Handler: &CoilsHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.compressor.stage2.switch",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Compressor Stage2",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Coil,
					Info:       "Compressor Stage2",
					Data: map[string]interface{}{
						"address": 0x2C,
						"width":   1,
						"type":    "b",
					},
				},
			},
			Handler: &CoilsHandler,
		},

		&sdk.Device{
			Kind:   "vem-plc.compressor.stage1.switch",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Compressor Stage2",
			Location: &sdk.Location{
				Rack:  "vem-location",
				Board: "vem-plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Coil,
					Info:       "Compressor Stage1",
					Data: map[string]interface{}{
						"address": 0x2D,
						"width":   1,
						"type":    "b",
					},
				},
			},
			Handler: &CoilsHandler,
		},
	}

	dumpDevices(t, registerDevices)
	dumpDevices(t, coilDevices)

	t.Logf("--- Mapping bulk read ---")
	bulkReadMapRegisters, keyOrderRegisters, err := MapBulkRead(registerDevices, true, false)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("bulkReadMapRegisters %#v", bulkReadMapRegisters)

	bulkReadMapCoils, keyOrderCoils, err := MapBulkRead(coilDevices, true, true)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("bulkReadMapCoils %#v", bulkReadMapCoils)
	t.Logf("--- Mapping bulk read end ---")

	dumpBulkReadMap(t, bulkReadMapRegisters, keyOrderRegisters)
	dumpBulkReadMap(t, bulkReadMapCoils, keyOrderCoils)

	// Validate the maps.
	// Registers
	if len(bulkReadMapRegisters) != 1 {
		t.Fatalf("Only one entry should be present, got %v", len(bulkReadMapRegisters))
	}

	expectedKey := ModbusBulkReadKey{
		Host:                 "10.193.4.250",
		Port:                 502,
		Timeout:              "10s",
		FailOnError:          false,
		MaximumRegisterCount: 0x7b,
	}

	readRegisters := bulkReadMapRegisters[expectedKey]
	if len(readRegisters) != 1 {
		t.Fatalf("Only one read should be required, got count %v, %#v", len(readRegisters), readRegisters)
	}

	readRegister := readRegisters[0]
	t.Logf("readRegister: %#v", readRegister)

	if readRegister.StartRegister != 0x01 {
		t.Fatalf("expected startRegister 0x01, got 0x%04x", readRegister.StartRegister)
	}

	if readRegister.RegisterCount != 0x40 {
		t.Fatalf("expected registerCount 0x40, got 0x%04x", readRegister.RegisterCount)
	}

	if len(readRegister.Devices) != 17 {
		t.Fatalf("expected 17 devices, got %v", len(readRegister.Devices))
	}

	// Coils
	if len(bulkReadMapCoils) != 1 {
		t.Fatalf("Only one entry should be present, got %v", len(bulkReadMapCoils))
	}

	readCoils := bulkReadMapCoils[expectedKey]
	if len(readCoils) != 1 {
		t.Fatalf("Only one read should be required, got count %v, %#v", len(readCoils), readCoils)
	}

	readCoil := readCoils[0]
	t.Logf("readCoil: %#v", readCoil)

	if readCoil.StartRegister != 0x03 {
		t.Fatalf("expected startRegister 0x01, got 0x%04x", readCoil.StartRegister)
	}

	if readCoil.RegisterCount != 0x2b {
		t.Fatalf("expected registerCount 0x2b, got 0x%04x", readCoil.RegisterCount)
	}

	if len(readCoil.Devices) != 8 {
		t.Fatalf("expected 8 devices, got %v", len(readCoil.Devices))
	}

	// Populate the maps to simulate readings and dump.
	// Holding Registers.
	populateBulkReadMap(t, bulkReadMapRegisters, keyOrderRegisters)
	dumpBulkReadMap(t, bulkReadMapRegisters, keyOrderRegisters)

	// Map the read data to the synse read contexts.
	readContextsRegisters, err := MapBulkReadData(bulkReadMapRegisters, keyOrderRegisters)
	if err != nil {
		t.Fatalf("Failed to map bulk read data, error: %v", err.Error())
	}
	dumpReadContexts(t, readContextsRegisters)

	// Verify read contexts and each reading.
	if len(readContextsRegisters) != 17 {
		t.Fatalf("expected 17 readContexts, got %v", len(readContextsRegisters))
	}

	if len(readContextsRegisters[0].Reading) != 1 {
		t.Fatalf("expected 1 reading in readContextsRegisters[0], got %v", len(readContextsRegisters[0].Reading))
	}

	expectedRegisterReadings := []*sdk.Reading{

		&sdk.Reading{
			Type:  "temperature",
			Info:  "HRC Mixed Fluid Temperature",
			Unit:  sdk.Unit{Name: "celsius", Symbol: "C"},
			Value: -17.72222222222222,
		},

		&sdk.Reading{
			Type:  "temperature",
			Info:  "Loop Entering Fluid Temperature",
			Unit:  sdk.Unit{Name: "celsius", Symbol: "C"},
			Value: 10.833333333333334,
		},

		&sdk.Reading{
			Type:  "flowGpm",
			Info:  "Minimum Flow Control Valve2 Feedback",
			Unit:  sdk.Unit{Name: "gallons per minute", Symbol: "gpm"},
			Value: int16(2057),
		},

		&sdk.Reading{
			Type:  "flowGpm",
			Info:  "System Fluid Flow",
			Unit:  sdk.Unit{Name: "gallons per minute", Symbol: "gpm"},
			Value: int16(2571),
		},

		&sdk.Reading{
			Type:  "InWCThousanths",
			Info:  "Server Rack Differential Pressure",
			Unit:  sdk.Unit{Name: "inches of water column", Symbol: "InWC"},
			Value: 3.085,
		},

		&sdk.Reading{
			Type:  "temperature",
			Info:  "System Leaving Fluid Temperature",
			Unit:  sdk.Unit{Name: "celsius", Symbol: "C"},
			Value: 239.27777777777777,
		},

		&sdk.Reading{
			Type:  "temperature",
			Info:  "Return Air Temperature",
			Unit:  sdk.Unit{Name: "celsius", Symbol: "C"},
			Value: 324.9444444444445,
		},

		&sdk.Reading{
			Type:  "temperature",
			Info:  "Outdoor Air Temperature",
			Unit:  sdk.Unit{Name: "celsius", Symbol: "C"},
			Value: 382.05555555555554,
		},

		&sdk.Reading{
			Type:  "temperature",
			Info:  "Cooling Coil Leaving Air Temperature",
			Unit:  sdk.Unit{Name: "celsius", Symbol: "C"},
			Value: 439.1666666666667,
		},

		&sdk.Reading{
			Type:  "psiTenths",
			Info:  "DX Discharge Gas Pressure",
			Unit:  sdk.Unit{Name: "pounds per square inch", Symbol: "psi"},
			Value: 1182.3,
		},

		&sdk.Reading{
			Type:  "temperature",
			Info:  "Return Air Temperature Setpoint",
			Unit:  sdk.Unit{Name: "celsius", Symbol: "C"},
			Value: 981.7222222222222,
		},

		&sdk.Reading{
			Type:  "fan_speed_percent",
			Info:  "HRF Speed Command",
			Unit:  sdk.Unit{Name: "percent", Symbol: "%"},
			Value: uint16(0x5455),
		},

		&sdk.Reading{
			Type:  "fan_speed_percent",
			Info:  "VEM Fan Speed Control",
			Unit:  sdk.Unit{Name: "percent", Symbol: "%"},
			Value: int16(22103),
		},

		&sdk.Reading{
			Type:  "flowGpmTenths",
			Info:  "Active Flow Setpoint",
			Unit:  sdk.Unit{Name: "gallons per minute", Symbol: "gpm"},
			Value: 2261.7000000000003,
		},

		&sdk.Reading{
			Type:  "fan_speed_percent",
			Info:  "VEM Fan Speed Actual",
			Unit:  sdk.Unit{Name: "percent", Symbol: "%"},
			Value: int16(25187),
		},

		&sdk.Reading{
			Type:  "flowGpmTenths",
			Info:  "Total System Flow",
			Unit:  sdk.Unit{Name: "gallons per minute", Symbol: "gpm"},
			Value: 2570.1000000000004,
		},

		&sdk.Reading{
			Type:  "fan_speed_percent_tenths",
			Info:  "VEM Fan Speed Minimum",
			Unit:  sdk.Unit{Name: "percent", Symbol: "%"},
			Value: 3238.3,
		},
	}
	t.Logf("expectedRegisterReadings: %#v", expectedRegisterReadings)

	// Get the actual readings in a slice. Verify readings are as expected.
	var actualRegisterReadings []*sdk.Reading
	for i := 0; i < len(readContextsRegisters); i++ {
		actualRegisterReadings = append(actualRegisterReadings, readContextsRegisters[i].Reading[0])
	}

	dumpReadings(t, actualRegisterReadings)
	verifyReadings(t, expectedRegisterReadings, actualRegisterReadings)

	// Coils
	populateBulkReadMap(t, bulkReadMapCoils, keyOrderCoils)
	dumpBulkReadMap(t, bulkReadMapCoils, keyOrderCoils)

	// Map the read data to the synse read contexts.
	readContextsCoils, err := MapBulkReadData(bulkReadMapCoils, keyOrderCoils)
	if err != nil {
		t.Fatalf("Failed to map bulk read data, error: %v", err.Error())
	}
	dumpReadContexts(t, readContextsCoils)

	// Verify read contexts and each reading.
	if len(readContextsCoils) != 8 {
		t.Fatalf("expected 8 readContexts, got %v", len(readContextsCoils))
	}

	if len(readContextsCoils[0].Reading) != 1 {
		t.Fatalf("expected 1 reading in readContextsCoils[0], got %v", len(readContextsCoils[0].Reading))
	}

	expectedCoilReadings := []*sdk.Reading{

		&sdk.Reading{
			Type:  "switch",
			Info:  "BMS Start",
			Unit:  sdk.Unit{Name: "", Symbol: ""},
			Value: false,
		},

		&sdk.Reading{
			Type:  "switch",
			Info:  "Compressor Bank A in Safety Shutdown",
			Unit:  sdk.Unit{Name: "", Symbol: ""},
			Value: false,
		},

		&sdk.Reading{
			Type:  "switch",
			Info:  "Compressor Bank B in Safety Shutdown",
			Unit:  sdk.Unit{Name: "", Symbol: ""},
			Value: false,
		},

		&sdk.Reading{
			Timestamp: "2019-01-25T02:40:25.062928076Z",
			Type:      "switch",
			Info:      "System Mode Stage3",
			Unit:      sdk.Unit{Name: "", Symbol: ""},
			Value:     true,
		},

		&sdk.Reading{
			Type:  "switch",
			Info:  "System Mode Stage2",
			Unit:  sdk.Unit{Name: "", Symbol: ""},
			Value: false,
		},

		&sdk.Reading{
			Type:  "switch",
			Info:  "BMS Keep Alive",
			Unit:  sdk.Unit{Name: "", Symbol: ""},
			Value: false,
		},

		&sdk.Reading{
			Type:  "switch",
			Info:  "Compressor Stage2",
			Unit:  sdk.Unit{Name: "", Symbol: ""},
			Value: false,
		},

		&sdk.Reading{
			Type:  "switch",
			Info:  "Compressor Stage1",
			Unit:  sdk.Unit{Name: "", Symbol: ""},
			Value: true,
		},
	}

	// Get the actual readings in a slice. Verify readings are as expected.
	var actualCoilReadings []*sdk.Reading
	for i := 0; i < len(readContextsCoils); i++ {
		actualCoilReadings = append(actualCoilReadings, readContextsCoils[i].Reading[0])
	}

	dumpReadings(t, actualCoilReadings)
	verifyReadings(t, expectedCoilReadings, actualCoilReadings)

	t.Logf("TestVEM end")
}

// Unable to connect to the device. Fail on error is false, which allows
// subsequent reads to potentially pass.
func TestReadHoldingRegisters_NoConnection(t *testing.T) {

	devices := []*sdk.Device{
		&sdk.Device{
			Kind:   "temperature",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Test Temperature",
			Location: &sdk.Location{
				Rack:  "location",
				Board: "board",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "1s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Temperature,
					Info:       "Test Temperature",
					Data: map[string]interface{}{
						"address": 0x01,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},
	}

	// Make the bulk read call.
	readContexts, err := bulkReadHoldingRegisters(devices)
	t.Logf("readContexts, len(readContexts), err: %#v, %v, %v", readContexts, len(readContexts), err)
	// With fail on error false, we should get a nil reading.
	if err != nil {
		t.Fatalf(err.Error())
	}
	verifySingleNilReading(t, readContexts)
}

// Unable to connect to the device. Fail on error is true, which fails all
// subsequent reads.
func TestReadHoldingRegisters_NoConnection_FailOnError(t *testing.T) {

	devices := []*sdk.Device{
		&sdk.Device{
			Kind:   "temperature",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Test Temperature",
			Location: &sdk.Location{
				Rack:  "location",
				Board: "board",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "1s",
				"failOnError": true,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Temperature,
					Info:       "Test Temperature",
					Data: map[string]interface{}{
						"address": 0x01,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},
	}

	// Make the bulk read call.
	readContexts, err := bulkReadHoldingRegisters(devices)
	t.Logf("readContexts, len(readContexts), err: %#v, %v, %v", readContexts, len(readContexts), err)
	// With fail on error true, we fail hard.
	if err == nil {
		t.Fatalf("Expected error, got nil")
	}
	// This is a good error message from goburrow/modbus that contains the ip and
	// port. Let's test that.
	// Possible errors:
	// dial tcp 10.193.4.250:502: i/o timeout
	// dial tcp 10.193.4.250:502: getsockopt: connection refused
	if !strings.Contains(err.Error(), "dial tcp 10.193.4.250:502") {
		t.Fatalf("Unexpected err: [%v]", err.Error())
	}
}

// Unable to connect to the device. Fail on error is false, which allows
// subsequent reads to potentially pass.
func TestReadInputRegisters_NoConnection(t *testing.T) {

	devices := []*sdk.Device{
		&sdk.Device{
			Kind:   "temperature",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Test Temperature",
			Location: &sdk.Location{
				Rack:  "location",
				Board: "board",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "1s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Temperature,
					Info:       "Test Temperature",
					Data: map[string]interface{}{
						"address": 0x01,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &InputRegisterHandler,
		},
	}

	// Make the bulk read call.
	readContexts, err := bulkReadInputRegisters(devices)
	t.Logf("readContexts, len(readContexts), err: %#v, %v, %v", readContexts, len(readContexts), err)
	// With fail on error false, we should get a nil reading.
	if err != nil {
		t.Fatalf(err.Error())
	}
	verifySingleNilReading(t, readContexts)
}

// Unable to connect to the device. Fail on error is false, which allows
// subsequent reads to potentially pass.
func TestReadCoils_NoConnection(t *testing.T) {

	devices := []*sdk.Device{
		&sdk.Device{
			Kind:   "switch",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Test Switch",
			Location: &sdk.Location{
				Rack:  "location",
				Board: "board",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "1s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Coil,
					Info:       "Test Switch",
					Data: map[string]interface{}{
						"address": 0x81,
						"width":   1,
						"type":    "b",
					},
				},
			},
			Handler: &CoilsHandler,
		},
	}

	// Make the bulk read call.
	readContexts, err := bulkReadCoils(devices)
	t.Logf("readContexts, len(readContexts), err: %#v, %v, %v", readContexts, len(readContexts), err)
	// With fail on error false, we should get a nil reading.
	if err != nil {
		t.Fatalf(err.Error())
	}
	verifySingleNilReading(t, readContexts)
}

// We will need a read for each device below due to different IPs.
func TestReadHoldingRegisters_MoreThanOneDevice_IP(t *testing.T) {

	devices := []*sdk.Device{

		&sdk.Device{
			Kind:   "pressure",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Pressure at IP Address 1",
			Location: &sdk.Location{
				Rack:  "location",
				Board: "plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.PsiTenths,
					Info:       "Pressure at IP Address 1",
					Data: map[string]interface{}{
						"address": 0x18,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "pressure",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Pressure at IP Address 2",
			Location: &sdk.Location{
				Rack:  "location",
				Board: "plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.251",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.PsiTenths,
					Info:       "Pressure at IP Address 2",
					Data: map[string]interface{}{
						"address": 0x18,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},
	}

	t.Logf("from the test: devices: %#v", devices)
	for i := 0; i < len(devices); i++ {
		t.Logf("test devices[%v]: %#v", i, devices[i])
	}

	t.Logf("--- Mapping bulk read ---")
	bulkReadMap, keyOrder, err := MapBulkRead(devices, true, false)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("bulkReadMap %#v", bulkReadMap)
	t.Logf("--- Mapping bulk read end ---")

	populateBulkReadMap(t, bulkReadMap, keyOrder)
	dumpBulkReadMap(t, bulkReadMap, keyOrder)

	// Map the read data to the synse read contexts.
	readContexts, err := MapBulkReadData(bulkReadMap, keyOrder)
	if err != nil {
		t.Fatalf("Failed to map bulk read data, error: %v", err.Error())
	}
	dumpReadContexts(t, readContexts)

	// Validate Map.
	// There should be two map entries because there are two different IP addresses.
	if len(bulkReadMap) != 2 {
		t.Fatalf("Two map entries should be present, got %v", len(bulkReadMap))
	}

	// Validate readings.
	expectedReadings := []*sdk.Reading{

		&sdk.Reading{
			Type:  "psiTenths",
			Info:  "Pressure at IP Address 1",
			Unit:  sdk.Unit{Name: "pounds per square inch", Symbol: "psi"},
			Value: .1,
		},

		&sdk.Reading{
			Type:  "psiTenths",
			Info:  "Pressure at IP Address 2",
			Unit:  sdk.Unit{Name: "pounds per square inch", Symbol: "psi"},
			Value: .1,
		},
	}
	t.Logf("expectedReadings: %#v", expectedReadings)

	var actualReadings []*sdk.Reading
	for i := 0; i < len(readContexts); i++ {
		actualReadings = append(actualReadings, readContexts[i].Reading[0])
	}

	dumpReadings(t, actualReadings)
	verifyReadings(t, expectedReadings, actualReadings)
}

// We will need a read for each device below due to different ports.
func TestReadHoldingRegisters_MoreThanOneDevice_Port(t *testing.T) {

	devices := []*sdk.Device{

		&sdk.Device{
			Kind:   "pressure",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Pressure at Port 502",
			Location: &sdk.Location{
				Rack:  "location",
				Board: "plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.PsiTenths,
					Info:       "Pressure at Port 502",
					Data: map[string]interface{}{
						"address": 0x18,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "pressure",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Pressure at Port 503",
			Location: &sdk.Location{
				Rack:  "location",
				Board: "plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        503,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.PsiTenths,
					Info:       "Pressure at Port 503",
					Data: map[string]interface{}{
						"address": 0x18,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},
	}

	t.Logf("from the test: devices: %#v", devices)
	for i := 0; i < len(devices); i++ {
		t.Logf("test devices[%v]: %#v", i, devices[i])
	}

	t.Logf("--- Mapping bulk read ---")
	bulkReadMap, keyOrder, err := MapBulkRead(devices, true, false)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("bulkReadMap %#v", bulkReadMap)
	t.Logf("--- Mapping bulk read end ---")

	populateBulkReadMap(t, bulkReadMap, keyOrder)
	dumpBulkReadMap(t, bulkReadMap, keyOrder)

	// Map the read data to the synse read contexts.
	readContexts, err := MapBulkReadData(bulkReadMap, keyOrder)
	if err != nil {
		t.Fatalf("Failed to map bulk read data, error: %v", err.Error())
	}
	dumpReadContexts(t, readContexts)

	// Validate Map.
	// There should be two map entries because this test requires two reads.
	if len(bulkReadMap) != 2 {
		t.Fatalf("Two map entries should be present, got %v", len(bulkReadMap))
	}

	// Validate readings.
	expectedReadings := []*sdk.Reading{

		&sdk.Reading{
			Type:  "psiTenths",
			Info:  "Pressure at Port 502",
			Unit:  sdk.Unit{Name: "pounds per square inch", Symbol: "psi"},
			Value: .1,
		},

		&sdk.Reading{
			Type:  "psiTenths",
			Info:  "Pressure at Port 503",
			Unit:  sdk.Unit{Name: "pounds per square inch", Symbol: "psi"},
			Value: .1,
		},
	}
	t.Logf("expectedReadings: %#v", expectedReadings)

	var actualReadings []*sdk.Reading
	for i := 0; i < len(readContexts); i++ {
		actualReadings = append(actualReadings, readContexts[i].Reading[0])
	}

	dumpReadings(t, actualReadings)
	verifyReadings(t, expectedReadings, actualReadings)
}

// We will need a read for each device below because we are spanning more
// registers than will fit in a single read.
func TestReadHoldingRegisters_MultipleReads000(t *testing.T) {

	devices := []*sdk.Device{

		&sdk.Device{
			Kind:   "pressure",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Pressure 1",
			Location: &sdk.Location{
				Rack:  "location",
				Board: "plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.PsiTenths,
					Info:       "Pressure 1",
					Data: map[string]interface{}{
						"address": 0x0,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "pressure",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Pressure 2",
			Location: &sdk.Location{
				Rack:  "location",
				Board: "plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.PsiTenths,
					Info:       "Pressure 2",
					Data: map[string]interface{}{
						"address": MaximumRegisterCount,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},
	}

	t.Logf("from the test: devices: %#v", devices)
	for i := 0; i < len(devices); i++ {
		t.Logf("test devices[%v]: %#v", i, devices[i])
	}

	t.Logf("--- Mapping bulk read ---")
	bulkReadMap, keyOrder, err := MapBulkRead(devices, true, false)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("bulkReadMap %#v", bulkReadMap)
	t.Logf("--- Mapping bulk read end ---")

	populateBulkReadMap(t, bulkReadMap, keyOrder)
	dumpBulkReadMap(t, bulkReadMap, keyOrder)

	// Map the read data to the synse read contexts.
	readContexts, err := MapBulkReadData(bulkReadMap, keyOrder)
	if err != nil {
		t.Fatalf("Failed to map bulk read data, error: %v", err.Error())
	}
	dumpReadContexts(t, readContexts)

	// Validate Map.
	// There should be one map entry with two reads because the registers are far
	// enough apart to span a single read.
	if len(bulkReadMap) != 1 {
		t.Fatalf("One map entry should be present, got %v", len(bulkReadMap))
	}

	// Validate two reads.
	expectedKey := ModbusBulkReadKey{
		Host:                 "10.193.4.250",
		Port:                 502,
		Timeout:              "10s",
		FailOnError:          false,
		MaximumRegisterCount: 0x7b,
	}

	if len(bulkReadMap[expectedKey]) != 2 {
		t.Fatalf("Expected two reads, got %v", len(bulkReadMap[expectedKey]))
	}

	// Validate readings.
	expectedReadings := []*sdk.Reading{

		&sdk.Reading{
			Type:  "psiTenths",
			Info:  "Pressure 1",
			Unit:  sdk.Unit{Name: "pounds per square inch", Symbol: "psi"},
			Value: .1,
		},

		&sdk.Reading{
			Type:  "psiTenths",
			Info:  "Pressure 2",
			Unit:  sdk.Unit{Name: "pounds per square inch", Symbol: "psi"},
			Value: .1,
		},
	}
	t.Logf("expectedReadings: %#v", expectedReadings)

	var actualReadings []*sdk.Reading
	for i := 0; i < len(readContexts); i++ {
		actualReadings = append(actualReadings, readContexts[i].Reading[0])
	}

	dumpReadings(t, actualReadings)
	verifyReadings(t, expectedReadings, actualReadings)
}

// We will need a read for each device below because we are spanning more
// registers than will fit in a single read for this test it is due to data
// width of the second register.
func TestReadHoldingRegisters_MultipleReads001(t *testing.T) {

	devices := []*sdk.Device{

		&sdk.Device{
			Kind:   "pressure",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Pressure 1",
			Location: &sdk.Location{
				Rack:  "location",
				Board: "plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.PsiTenths,
					Info:       "Pressure",
					Data: map[string]interface{}{
						"address": 0x0,
						"width":   1,
						"type":    "s16",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},

		&sdk.Device{
			Kind:   "temperature",
			Plugin: "synse-modbus-ip-plugin",
			Info:   "Temperature",
			Location: &sdk.Location{
				Rack:  "location",
				Board: "plc",
			},
			Data: map[string]interface{}{
				"host":        "10.193.4.250",
				"port":        502,
				"timeout":     "10s",
				"failOnError": false,
			},
			Outputs: []*sdk.Output{
				&sdk.Output{
					OutputType: outputs.Temperature,
					Info:       "Temperature",
					Data: map[string]interface{}{
						"address": MaximumRegisterCount - 1,
						"width":   2,
						"type":    "s32",
					},
				},
			},
			Handler: &HoldingRegisterHandler,
		},
	}

	t.Logf("from the test: devices: %#v", devices)
	for i := 0; i < len(devices); i++ {
		t.Logf("test devices[%v]: %#v", i, devices[i])
	}

	t.Logf("--- Mapping bulk read ---")
	bulkReadMap, keyOrder, err := MapBulkRead(devices, true, false)
	if err != nil {
		t.Fatalf(err.Error())
	}
	t.Logf("bulkReadMap %#v", bulkReadMap)
	t.Logf("--- Mapping bulk read end ---")

	populateBulkReadMap(t, bulkReadMap, keyOrder)
	dumpBulkReadMap(t, bulkReadMap, keyOrder)

	// Map the read data to the synse read contexts.
	readContexts, err := MapBulkReadData(bulkReadMap, keyOrder)
	if err != nil {
		t.Fatalf("Failed to map bulk read data, error: %v", err.Error())
	}
	dumpReadContexts(t, readContexts)

	// Validate Map.
	// There should be one map entry with two reads because the registers are far
	// enough apart to span a single bulk read.
	if len(bulkReadMap) != 1 {
		t.Fatalf("One map entry should be present, got %v", len(bulkReadMap))
	}

	// Validate readings.
	expectedReadings := []*sdk.Reading{

		&sdk.Reading{
			Type:  "psiTenths",
			Info:  "Pressure",
			Unit:  sdk.Unit{Name: "pounds per square inch", Symbol: "psi"},
			Value: .1,
		},

		&sdk.Reading{
			Type:  "temperature",
			Info:  "Temperature",
			Unit:  sdk.Unit{Name: "celsius", Symbol: "C"},
			Value: 3651.722222222222,
		},
	}

	// Validate two reads.
	expectedKey := ModbusBulkReadKey{
		Host:                 "10.193.4.250",
		Port:                 502,
		Timeout:              "10s",
		FailOnError:          false,
		MaximumRegisterCount: 0x7b,
	}

	if len(bulkReadMap[expectedKey]) != 2 {
		t.Fatalf("Expected two reads, got %v", len(bulkReadMap[expectedKey]))
	}

	t.Logf("expectedReadings: %#v", expectedReadings)

	var actualReadings []*sdk.Reading
	for i := 0; i < len(readContexts); i++ {
		actualReadings = append(actualReadings, readContexts[i].Reading[0])
	}

	dumpReadings(t, actualReadings)
	verifyReadings(t, expectedReadings, actualReadings)
}

// TODO:
// No read data. (Probably the same as no connection)
// Insufficient data.
// Write connection failures.
// Additional VEM devices. 6 e-gauges. Carousel.
