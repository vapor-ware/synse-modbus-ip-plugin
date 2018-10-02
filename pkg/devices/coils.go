package devices

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/utils"
	"github.com/vapor-ware/synse-sdk/sdk"
)

// CoilsHandler is a handler that should be used for all devices/outputs
// that read from/write to coils.
var CoilsHandler = sdk.DeviceHandler{
	Name: "coils",
	Read: readCoils,
}

// readCoil is the read function for the coils device handler.
func readCoils(device *sdk.Device) ([]*sdk.Reading, error) {
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

	// For each device instance, we will have various outputs defined.
	// The outputs here should contain their own data which tells us what
	// the register address and the read width are.
	var readings []*sdk.Reading
	for i, output := range device.Outputs {
		log.Debugf(" -- [%d] ----------", i)
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

		// Use the output data config to get the coils reading
		results, err := client.ReadCoils(uint16(outputData.Address), uint16(outputData.Width))
		if err != nil {
			if failOnErr {
				return nil, err
			}
			log.Errorf("failed to read coils for output %v: %v", outputData, err)
			continue
		}

		// Cast the raw reading value to the specified output type
		data, err := utils.CastToType(outputData.Type, results)
		if err != nil {
			if failOnErr {
				return nil, err
			}
			log.Errorf("error casting reading data: %v", err)
			continue
		}
		log.Debugf("coils read result: %v", data)

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
