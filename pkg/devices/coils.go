package devices

import (
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
	"github.com/vapor-ware/synse-sdk/sdk"
)

// CoilsHandler is a handler that should be used for all devices/outputs
// that read from/write to coils.
var CoilsHandler = sdk.DeviceHandler{
	Name:  "coil",
	Read:  readCoils,
	Write: writeCoils,
}

// readCoils is the read function for the coils device handler.
func readCoils(device *sdk.Device) ([]*sdk.Reading, error) {
	if device == nil {
		return nil, fmt.Errorf("readCoils device is nil")
	}

	modbusConfig, client, err := GetModbusClientAndConfig(device)
	if err != nil {
		return nil, err
	}

	failOnErr := (*modbusConfig).FailOnError
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
		results, err := (*client).ReadCoils(uint16(outputData.Address), uint16(outputData.Width))
		if err != nil {
			if failOnErr {
				return nil, err
			}
			log.Errorf("failed to read coils for output %v: %v", outputData, err)
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

// writeCoils is the read function for the coils device handler.
func writeCoils(device *sdk.Device, data *sdk.WriteData) (err error) {

	if device == nil {
		return fmt.Errorf("device is nil")
	}
	if data == nil {
		return fmt.Errorf("data is nil")
	}

	_, client, err := GetModbusClientAndConfig(device)
	if err != nil {
		return err
	}

	// Pull out the data to send on the wire from data.Data.
	modbusData := data.Data

	output, err := GetOutput(device)
	if err != nil {
		return err
	}

	// Translate the data. For whatever reason, the modbus interface wants 0
	// for false and FF00 for true.
	dataString := string(modbusData)
	var coilData uint16
	switch dataString {
	case "0", "false", "False":
		coilData = 0
	case "1", "true", "True":
		coilData = 0xFF00
	default:
		return fmt.Errorf("unknown coil data %v", coilData)
	}

	register := (*output).Data["address"]
	registerInt, ok := register.(int)
	if !ok {
		return fmt.Errorf("Unable to convert (*output).Data[address] to uint16: %v", (*output).Data["address"])
	}
	registerUint16 := uint16(registerInt)

	_, err = (*client).WriteSingleCoil(registerUint16, coilData)
	return err
}
