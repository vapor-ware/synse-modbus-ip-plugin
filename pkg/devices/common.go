package devices

// This file contains common modbus device code.
import (
	"fmt"

	"github.com/goburrow/modbus"
	"github.com/mitchellh/mapstructure"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/utils"
	"github.com/vapor-ware/synse-sdk/sdk"
)

// GetModbusClientAndConfig is common code to get the modbus configuration and client from the device configuration.
func GetModbusClientAndConfig(device *sdk.Device) (modbusConfig *config.ModbusDeviceData, client *modbus.Client, err error) {

	// Pull the modbus configuration out of the device Data fields.
	var deviceData config.ModbusDeviceData
	err = mapstructure.Decode(device.Data, &deviceData)
	if err != nil {
		return nil, nil, err
	}

	// Create the modbus client from the configuration data.
	cli, err := utils.NewClient(&deviceData)
	if err != nil {
		return nil, nil, err
	}
	return &deviceData, &cli, nil
}

// GetOutput is a helper to get the first output for a device. Called by Write functions.
func GetOutput(device *sdk.Device) (output *sdk.Output, err error) {

	if device == nil {
		return nil, fmt.Errorf("device is nil")
	}
	if device.Outputs == nil {
		return nil, fmt.Errorf("device.Outputs is nil")
	}

	// Checking for one output for now. We need the register to write.
	// We could support multiple outputs in the future by matching against
	// info if we like.
	length := len(device.Outputs)
	if length == 0 {
		return nil, fmt.Errorf("no device outputs")
	}

	if length > 1 {
		return nil, fmt.Errorf("multiple device outputs unsupported")
	}
	return device.Outputs[0], nil
}

// UnpackReading is a wrapper for CastToType and MakeReading.
func UnpackReading(output *sdk.Output, typeName string, rawReading []byte, failOnErr bool) (reading *sdk.Reading, err error) {

	// Cast the raw reading value to the specified output type
	data, err := utils.CastToType(typeName, rawReading)
	if err != nil {
		if failOnErr {
			return nil, err
		}
		return nil, nil // No reading.
	}

	reading, err = output.MakeReading(data)
	if err != nil {
		// In this case we will not check the 'failOnError' flag because
		// this isn't an issue with reading the device, its a configuration
		// issue and the error should be noted.
		return nil, err
	}
	return
}
