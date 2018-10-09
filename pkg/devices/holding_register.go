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

		// Handle Divisor.
		log.Debugf("outputData.Divisor: %T, %v", outputData.Divisor, outputData.Divisor)
		if outputData.Divisor != 0 && outputData.Divisor != 1 {
			log.Debugf("data: %T, %v", data, data)
			var floatData float64
			switch data.(type) {

			// uint
			case uint16:
				uint16Data := data.(uint16)
				floatData = float64(uint16Data)
			case uint32:
				uint32Data := data.(uint32)
				floatData = float64(uint32Data)
			case uint64:
				uint64Data := data.(uint64)
				floatData = float64(uint64Data)

			// int
			case int16:
				int16Data := data.(int16)
				floatData = float64(int16Data)
			case int32:
				int32Data := data.(int32)
				floatData = float64(int32Data)
			case int64:
				int64Data := data.(int64)
				floatData = float64(int64Data)

			// float
			case float32:
				float32Data := data.(float32)
				floatData = float64(float32Data)
			case float64:
				float64Data := data.(float64)
				floatData = float64Data

			// NYI
			default:
				err = fmt.Errorf("Unable to convert data 0x%0x to int", data)
				if failOnErr {
					return nil, err
				}
				log.Errorf(err.Error())
				continue
			}

			data = floatData / outputData.Divisor
		}
		log.Debugf("data after divisor: %T, %v", data, data)

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
		log.Debugf("Appending reading successfully")
		readings = append(readings, reading)
	}

	log.Debugf("readHoldingRegister end, readings: %+v", readings)
	return readings, nil
}
