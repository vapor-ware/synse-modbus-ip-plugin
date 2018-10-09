package devices

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/utils"
	"github.com/vapor-ware/synse-sdk/sdk"
)

// HoldingRegisterHandler is a handler which should be used for all devices/outputs
// that read from/write to holding registers.
var HoldingRegisterHandler = sdk.DeviceHandler{
	Name: "holding_register",
	Read: readHoldingRegister,
}

// readHoldingRegister is the read function for the holding register device handler.
func readHoldingRegister(device *sdk.Device) ([]*sdk.Reading, error) {

	// FIXME (etd) - holding registers, coils, and input registers all do pretty much
	// the same thing on read here.. consider abstracting this out so all we have to do
	// is something along the lines of:
	//
	//   func readHoldingRegister(device *sdk.Device) ([]*sdk.Reading, error) {
	//      return utils.Read(device, "holding")
	//   }
	log.Debugf("readHoldingRegister start: %+v", device)

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
		log.Debugf("  Device OutputType:  %v", output.OutputType)
		log.Debugf("  Device Output Data: %v", output.Data)

		// Get the output data config
		var outputData config.ModbusOutputData
		err := mapstructure.Decode(output.Data, &outputData)
		if err != nil {
			if failOnErr {
				return nil, err
			}
			log.Errorf("failed to parse output config: %v", err)
			continue
		}

		// Now use that to get the holding register reading
		log.Debugf(
			"Begin Reading holding register address 0x%0x, width 0x%x",
			uint16(outputData.Address),
			uint16(outputData.Width))

		results, err := client.ReadHoldingRegisters(
			uint16(outputData.Address), uint16(outputData.Width))
		if err != nil {
			if failOnErr {
				return nil, err
			}
			log.Errorf("failed to read holding registers for output %v: %v", outputData, err)
			continue
		}

		log.Debugf("ReadHoldingRegisters: results: 0x%0x, len(results) 0x%0x", results, len(results))
		// Cast the raw reading value to the specified output type
		data, err := utils.CastToType(outputData.Type, results)
		if err != nil {
			if failOnErr {
				return nil, err
			}
			log.Errorf("error casting reading data: %v, error %v", data, err)
			continue
		}

		log.Debugf("holding register read result: %T, %v", data, data)

		// Handle English to Metric.
		conversion, present := output.Data["conversion"]
		if present {
			// This is currently the only supported conversion.
			if conversion == "englishToMetric" {
				data, err = utils.ConvertEnglishToMetric(output.OutputType.Name, data)
				if err != nil {
					if failOnErr {
						return nil, err
					}
					log.Errorf(err.Error())
					continue
				}
				log.Debugf("data after english to metric conversion: %T, %v", data, data)
			} else {
				err = fmt.Errorf("Unsupported conversion key in configuration: %v", conversion)
				if failOnErr {
					return nil, err
				}
				log.Errorf(err.Error())
				continue
			}
		}

		reading, err := output.MakeReading(data)
		if err != nil {
			// In this case we will not check the 'failOnError' flag because
			// this isn't an issue with reading the device, its a configuration
			// issue and the error should be noted.
			return nil, err
		}
		readings = append(readings, reading)
	}

	log.Debugf("readHoldingRegister end, readings: %+v", readings)
	return readings, nil
}
