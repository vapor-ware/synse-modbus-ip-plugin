package devices

import (
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/protocol"
	"github.com/vapor-ware/synse-sdk/sdk"
	"fmt"
	"strings"
	log "github.com/Sirupsen/logrus"
)


var InputRegisterHandler = sdk.DeviceHandler{
	Name: "input_register",
	Read: readInputRegister,
}

func readInputRegister(device *sdk.Device) ([]*sdk.Reading, error) {
	client, err := protocol.NewClient(device.Data)
	if err != nil {
		return nil, err
	}

	var readings []*sdk.Reading

	// For each device instance, we will have various outputs defined.
	// The outputs here should contain their own data that tells us what
	// the register address and read width are.
	for _, output := range device.Outputs {
		log.Info("Device Output Data: %#v", output.Data)
		addr, ok := output.Data["address"]
		if !ok {
			return nil, fmt.Errorf("output data 'address' not specified, but required")
		}
		address, ok := addr.(int)
		if !ok {
			return nil, fmt.Errorf("output data 'address' (%d) should be uint16 but is %T", address, address)
		}

		wdth, ok := output.Data["width"]
		if !ok {
			return nil, fmt.Errorf("output data 'width' not specified, but required")
		}
		width, ok := wdth.(int)
		if !ok {
			return nil, fmt.Errorf("output data 'width' (%d) should be uint16 but is %T", width, width)
		}

		// Now use that to get the reading
		results, err := client.ReadInputRegisters(uint16(address), uint16(width))
		if err != nil {
			return nil, err
		}

		// We want to convert the data to the appropriate type
		dt, ok := output.Data["type"]
		if !ok {
			return nil, fmt.Errorf("output data 'type' not specified, but required")
		}
		dataType, ok := dt.(string)
		if !ok {
			return nil, fmt.Errorf("output data 'type' (%d) should be string but is %T", dataType, dataType)
		}

		var data interface{}
		switch strings.ToLower(dataType) {
		case "u32", "uint32":
			// unsigned 32-bit integer
			data = protocol.Bytes(results).Uint32()
		case "u64", "uint64":
			// unsigned 64-bit integer
			data = protocol.Bytes(results).Uint64()
		case "s32", "int32":
			// signed 32-bit integer
			data, err = protocol.Bytes(results).Int32()
			if err != nil {
				return nil, err
			}
		case "s64", "int64":
			// signed 63-bit integer
			data, err = protocol.Bytes(results).Int64()
			if err != nil {
				return nil, err
			}
		case "f32", "float32":
			// 32-bit floating point number
			data = protocol.Bytes(results).Float32()
		case "f64", "float64":
			// 64-bit floating point number
			data = protocol.Bytes(results).Float64()
		default:
			// The type is not supported. This could be a typo, or it could be a
			// new type that needs to be added in.
			return nil, fmt.Errorf("output data 'type' specifies unsupported type '%s'", dataType)
		}

		readings = append(readings, output.MakeReading(data))
	}
	return readings, nil
}

/*
NOTE:

For the first rev of this plugin, we have the device handler defined
here explicitly. Once the SDK is updated to better support generalized
plugin structures, this will be updated and specific device handlers
will be replaced with a generalized handler.
*/

// EG4115PowerMeter is the handler for the eGauge 4115 Power Meter device.
var EG4115PowerMeter = sdk.DeviceHandler{
	Name: "power",
	Read: readEG4115PowerMeter,
}

func readEG4115PowerMeter(device *sdk.Device) ([]*sdk.Reading, error) {

	client, err := protocol.NewClient(device.Data)
	if err != nil {
		return nil, err
	}

	// FIXME (etd) - for now, the handling for the EG4115 power meter device is
	// going to be completely hardcoded. this plugin is not generalizable yet, as
	// there are many changes that need to happen in the next rev of the SDK first.
	// This gets things working for the power meter, but thats just about it.
	results, err := client.ReadInputRegisters(500, 8)
	if err != nil {
		return nil, err
	}

	l1RMS := protocol.Float32FromBytes(results[0:4])
	l2RMS := protocol.Float32FromBytes(results[4:8])
	l1l2RMS := protocol.Float32FromBytes(results[12:16])

	readings := []*sdk.Reading{
		device.GetOutput("voltage").MakeReading(l1RMS),
		device.GetOutput("voltage").MakeReading(l2RMS),
		device.GetOutput("voltage").MakeReading(l1l2RMS),
	}
	return readings, nil
}
