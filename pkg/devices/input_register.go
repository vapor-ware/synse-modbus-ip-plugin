package devices

import (
	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
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

	modbusConfig, client, err := GetModbusClientAndConfig(device)
	if err != nil {
		return nil, err
	}

	failOnErr := (*modbusConfig).FailOnError
	var readings []*sdk.Reading

	// For each device instance, we will have various outputs defined.
	// The outputs here should contain their own data that tells us what
	// the register address and read width are.
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

		// Now use that to get the input register reading
		results, err := (*client).ReadInputRegisters(uint16(outputData.Address), uint16(outputData.Width))
		if err != nil {
			if failOnErr {
				return nil, err
			}
			log.Errorf("failed to read input registers for output %v: %v", outputData, err)
			continue
		}

		reading, err := UnpackReading(output, outputData.Type, results, failOnErr)
		if err != nil {
			return nil, err
		}
		readings = append(readings, reading)
	}
	return readings, nil
}
