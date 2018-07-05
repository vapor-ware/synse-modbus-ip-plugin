package devices

import (
	"fmt"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
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
	var deviceData config.ModbusDeviceData
	err := mapstructure.Decode(device.Data, &deviceData)
	if err != nil {
		return nil, err
	}

	client, err := utils.NewClient(&deviceData)
	if err != nil {
		return nil, err
	}

	failOnErr := deviceData.FailOnError
	log.Debugf("fail on error: %v", failOnErr)

	var readings []*sdk.Reading

	// For each device instance, we will have various outputs defined.
	// The outputs here should contain their own data that tells us what
	// the register address and read width are.
	for i, output := range device.Outputs {
		log.Debugf(" -- [%d] ----------", i)
		log.Debugf("  Device Output Data: %v", output.Data)

		var outputData config.ModbusOutputData
		err := mapstructure.Decode(output.Data, &outputData)
		if err != nil {
			if failOnErr {
				return nil, err
			}
			log.Errorf("failed to parse output config: %v", err)
			continue
		}

		// Now use that to get the reading
		results, err := client.ReadInputRegisters(uint16(outputData.Address), uint16(outputData.Width))
		if err != nil {
			if failOnErr {
				return nil, err
			}
			log.Errorf("failed to read input registers for output %v: %v", outputData, err)
			continue
		}

		var data interface{}
		switch strings.ToLower(outputData.Type) {
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
				if failOnErr {
					return nil, err
				}
				log.Errorf("failed to cast bytes (%v) to int32: %v", results, err)
				continue
			}
		case "s64", "int64":
			// signed 64-bit integer
			data, err = utils.Bytes(results).Int64()
			if err != nil {
				if failOnErr {
					return nil, err
				}
				log.Errorf("failed to cast bytes (%v) to int64: %v", results, err)
				continue
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
			if failOnErr {
				return nil, fmt.Errorf("output data 'type' specifies unsupported type '%s'", outputData.Type)
			}
			log.Errorf("output data 'type' is unsupported: %s", outputData.Type)
			continue
		}
		log.Debugf("  result: %v", data)

		reading, err := output.MakeReading(data)
		if err != nil {
			// In this case we will not check the 'failOnError' flag because
			// this isn't an issue with reading the device, its a configuration
			// issue and the error should be noted.
			return nil, err
		}
		readings = append(readings, reading)
	}
	return readings, nil
}
