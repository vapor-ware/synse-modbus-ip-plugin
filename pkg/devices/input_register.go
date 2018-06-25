package devices

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/utils"
	"github.com/vapor-ware/synse-sdk/sdk"
)

// InputRegisterHandler is a handler that should be used for all devices/outputs
// that read input registers.
var InputRegisterHandler = sdk.DeviceHandler{
	Name: "input_register",
	Read: readInputRegister,
}

// readInputRegister is the read function for the input register device handler.
func readInputRegister(device *sdk.Device) ([]*sdk.Reading, error) {
	client, err := utils.NewClient(device.Data)
	if err != nil {
		return nil, err
	}

	var readings []*sdk.Reading

	// For each device instance, we will have various outputs defined.
	// The outputs here should contain their own data that tells us what
	// the register address and read width are.
	for i, output := range device.Outputs {
		log.Debugf(" -- [%d] ----------", i)
		log.Debugf("  Device Output Data: %v", output.Data)
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
			// FIXME (etd): Should we fail here? We are reading multiple registers in
			// this loop. If even one fails, that will fail the read for every register.
			// I think what would be better is to just log this error out and move on.
			// We could also track the number of errors we got when reading. Then, if
			// we failed to read from all registers (or some % of them?) then we can
			// return an error.
			//
			// On the other hand, all of these registers are on the same device, so if
			// a few are failing, that could mean something is wrong with the device, or
			// the device is mis-configured, in which case we probably do want this to
			// error out.
			return nil, err
		}

		// We want to convert the data to the appropriate type
		dt, ok := output.Data["type"]
		if !ok {
			return nil, fmt.Errorf("output data 'type' not specified, but required")
		}
		dataType, ok := dt.(string)
		if !ok {
			return nil, fmt.Errorf("output data 'type' (%s) should be string but is %T", dataType, dataType)
		}

		var data interface{}
		switch strings.ToLower(dataType) {
		case "u32", "uint32":
			// unsigned 32-bit integer
			data = utils.Bytes(results).Uint32()
		case "u64", "uint64":
			// unsigned 64-bit integer
			data = utils.Bytes(results).Uint64()
		case "s32", "int32":
			// signed 32-bit integer
			data, err = utils.Bytes(results).Int32()
			if err != nil {
				return nil, err
			}
		case "s64", "int64":
			// signed 63-bit integer
			data, err = utils.Bytes(results).Int64()
			if err != nil {
				return nil, err
			}
		case "f32", "float32":
			// 32-bit floating point number
			data = utils.Bytes(results).Float32()
		case "f64", "float64":
			// 64-bit floating point number
			data = utils.Bytes(results).Float64()
		default:
			// The type is not supported. This could be a typo, or it could be a
			// new type that needs to be added in.
			return nil, fmt.Errorf("output data 'type' specifies unsupported type '%s'", dataType)
		}

		log.Debugf("  result: %v", data)
		readings = append(readings, output.MakeReading(data))
	}
	return readings, nil
}
