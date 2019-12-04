package devices

import (
	"errors"
	"fmt"
	"strings"

	"github.com/goburrow/modbus"
	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/output"
)

const handlerCoil = "coil"

// CoilsHandler is a handler that should be used for all devices/outputs
// that read from/write to coils.
var CoilsHandler = sdk.DeviceHandler{
	Name: handlerCoil,
	BulkRead: func(devices []*sdk.Device) (contexts []*sdk.ReadContext, e error) {
		managers, found := DeviceManagers[handlerCoil]
		if !found {
			return nil, errors.New("no device manager(s) found for coil handler")
		}
		return bulkReadCoils(managers)
	},
	Write: func(device *sdk.Device, data *sdk.WriteData) error {
		if device == nil {
			return fmt.Errorf("unable to write to coil: device is nil")
		}
		if data == nil {
			return fmt.Errorf("unable to write to coil: data is nil")
		}
		register, ok := device.Data["address"].(int)
		if !ok {
			return fmt.Errorf("unable to convert device data address (%v) to int", device.Data["address"])
		}

		client, err := NewModbusClient(device)
		if err != nil {
			return err
		}

		return writeCoil(client, uint16(register), data)
	},
}

// writeCoil validates and writes the provided data to the provided modbus register
// for the given client.
//
// This is broken apart from the device handler write function to make it easier to test.
func writeCoil(client modbus.Client, register uint16, data *sdk.WriteData) error {
	// Translate the configured coil data into a format accepted
	// by the modbus client.
	coilData, err := getCoilData(data.Data)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"address": register,
		"data":    coilData,
	}).Debug("writing to coil")
	_, err = client.WriteSingleCoil(register, coilData)
	return err
}

func bulkReadCoils(managers []*ModbusDeviceManager) ([]*sdk.ReadContext, error) {
	var readings []*sdk.ReadContext

	for _, manager := range managers {
		err := manager.ParseBlocks()
		if err != nil {
			return nil, err
		}
		for _, block := range manager.Blocks {
			// Perform the bulk read on the register block.
			results, err := manager.Client.ReadCoils(block.StartRegister, block.RegisterCount)
			if err != nil {
				if manager.FailOnError {
					return nil, err
				}
				results = []byte{}
			}

			/// Trim the result bytes to the expected length of bytes as per the calculated
			// register count. This ensures there are no padded values included in the data
			// for subsequent processing.
			if len(results) > 0 {
				results = results[0 : 2*block.RegisterCount] // Two bytes per register TODO double check this
			}
			block.Results = results

			// Parse the results from the bulk read. This will create the readings for
			// each device.
			for _, device := range block.Devices {
				out := output.Get(device.Device.Output)

				reading, err := UnpackCoilReading(out, block, device)
				if err != nil {
					return nil, err
				}

				readCtx := sdk.NewReadContext(device.Device, []*output.Reading{reading})
				readings = append(readings, readCtx)
			}
		}
	}

	return readings, nil
}

// getCoilData translates the device write data for a modbus coil from the configured
// byte array to an integer. The modbus interface wants 0 for false and ff00 for true.
func getCoilData(data []byte) (uint16, error) {
	switch strings.ToLower(string(data)) {
	case "0", "false":
		return 0, nil
	case "1", "true":
		return 0xff00, nil
	default:
		return 0, fmt.Errorf("unexpected coil data: %v", data)
	}
}
