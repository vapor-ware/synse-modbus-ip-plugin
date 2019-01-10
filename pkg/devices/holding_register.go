package devices

import (
	"fmt"
	"strconv"

	log "github.com/Sirupsen/logrus"
	"github.com/mitchellh/mapstructure"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
	"github.com/vapor-ware/synse-sdk/sdk"
)

// HoldingRegisterHandler is a handler which should be used for all devices/outputs
// that read from/write to holding registers.
var HoldingRegisterHandler = sdk.DeviceHandler{
	Name:  "holding_register",
	Read:  readHoldingRegister,
	Write: writeHoldingRegister,
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

	modbusConfig, client, err := GetModbusClientAndConfig(device)
	if err != nil {
		return nil, err
	}

	failOnErr := (*modbusConfig).FailOnError
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

		results, err := (*client).ReadHoldingRegisters(
			uint16(outputData.Address), uint16(outputData.Width))
		if err != nil {
			if failOnErr {
				return nil, err
			}
			log.Errorf("failed to read holding registers for output %v: %v", outputData, err)
			continue
		}

		log.Debugf("ReadHoldingRegisters: results: 0x%0x, len(results) 0x%0x", results, len(results))

		reading, err := UnpackReading(output, outputData.Type, results, failOnErr)
		if err != nil {
			return nil, err
		}
		readings = append(readings, reading)
	}

	log.Debugf("readHoldingRegister end, readings: %+v", readings)
	return readings, nil
}

// writeHoldingRegister is the write function for the holding register device handler.
func writeHoldingRegister(device *sdk.Device, data *sdk.WriteData) (err error) {

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

	// Translate the data. This is currently a hex string.
	dataString := string(modbusData)
	register64, err := strconv.ParseUint(dataString, 16, 16)
	if err != nil {
		return fmt.Errorf("Unable to parse uint16 %v", dataString)
	}
	registerData := uint16(register64)

	register := (*output).Data["address"]
	registerInt, ok := register.(int)
	if !ok {
		return fmt.Errorf("Unable to convert (*output).Data[address] to uint16: %v", (*output).Data["address"])
	}
	registerUint16 := uint16(registerInt)

	_, err = (*client).WriteSingleRegister(registerUint16, registerData)
	return err
}
