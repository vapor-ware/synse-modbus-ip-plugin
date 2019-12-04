package devices

import (
	"fmt"
	"strconv"

	"github.com/goburrow/modbus"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/output"
)

const handlerHoldingRegister = "holding_register"

// HoldingRegisterHandler is a handler which should be used for all devices/outputs
// that read from/write to holding registers.
var HoldingRegisterHandler = sdk.DeviceHandler{
	Name: handlerHoldingRegister,
	BulkRead: func(devices []*sdk.Device) (contexts []*sdk.ReadContext, e error) {
		managers, found := DeviceManagers[handlerHoldingRegister]
		if !found {
			return nil, errors.New("no device manager(s) found for holding register handler")
		}
		return bulkReadHoldingRegisters(managers)
	},
	Write: func(device *sdk.Device, data *sdk.WriteData) error {
		if device == nil {
			return fmt.Errorf("unable to write to holding register: device is nil")
		}
		if data == nil {
			return fmt.Errorf("unable to write to holding register: data is nil")
		}
		register, ok := device.Data["address"].(int)
		if !ok {
			return fmt.Errorf("unable to convert device data address (%v) to int", device.Data["address"])
		}

		client, err := NewModbusClient(device)
		if err != nil {
			return err
		}

		return writeHoldingRegister(client, uint16(register), data)
	},
}

func bulkReadHoldingRegisters(managers []*ModbusDeviceManager) ([]*sdk.ReadContext, error) {
	var readings []*sdk.ReadContext

	for _, manager := range managers {
		err := manager.ParseBlocks()
		if err != nil {
			return nil, err
		}
		for _, block := range manager.Blocks {

			results, err := manager.Client.ReadHoldingRegisters(block.StartRegister, block.RegisterCount)
			if err != nil {
				if manager.FailOnError {
					return nil, err
				}
				results = []byte{}
			}

			// Trim the result bytes to the expected length of bytes as per the calculated
			// register count. This ensures there are no padded values included in the data
			// for subsequent processing.
			if len(results) > 0 {
				results = results[0 : 2*block.RegisterCount] // Two bytes per register
			}
			block.Results = results

			// Parse the results from the bulk read. This will create the readings
			// for each device.
			for _, device := range block.Devices {
				out := output.Get(device.Device.Output)

				reading, err := UnpackRegisterReading(out, block, device)
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

func writeHoldingRegister(client modbus.Client, register uint16, data *sdk.WriteData) error {
	// Translate the configured holding register data into a format accepted
	// by the modbus client.
	registerData, err := getHoldingRegisterData(data.Data)
	if err != nil {
		return err
	}

	log.WithFields(log.Fields{
		"address": register,
		"data":    registerData,
	}).Debug("writing to holding register")
	_, err = client.WriteSingleRegister(register, registerData)
	return err
}

// getHoldingRegisterData translates the device write data for a modbus holding register from
// the byte array to an integer. The register data comes in as a hex string.
func getHoldingRegisterData(data []byte) (uint16, error) {
	r, err := strconv.ParseUint(string(data), 16, 16)
	if err != nil {
		return 0, errors.Wrapf(err, "unable to parse to uint16: %v", data)
	}
	return uint16(r), nil
}
