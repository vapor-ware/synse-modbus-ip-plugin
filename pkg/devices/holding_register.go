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

// HoldingRegisterHandler is a device handler used to read from and write to modbus
// holding registers.
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

// bulkReadHoldingRegisters performs a bulk read for modbus devices configured to use the holding
// register handler. It gets those devices from the device manager(s) associated with the handler,
// which is populated on plugin startup.
func bulkReadHoldingRegisters(managers []*ModbusDeviceManager) ([]*sdk.ReadContext, error) {
	var readings []*sdk.ReadContext

	log.Debug("starting bulk read for holding registers")
	for _, manager := range managers {
		// Attempt to parse the manager's devices into register blocks for bulk read
		// if they have not already been parsed.
		if err := manager.ParseBlocks(); err != nil {
			log.WithError(err).Error("failed to parse devices into read blocks")
			return nil, err
		}

		for _, block := range manager.Blocks {
			log.WithFields(log.Fields{
				"startRegister": block.StartRegister,
				"registerCount": block.RegisterCount,
			}).Debug("reading holding registers for block")
			results, err := manager.Client.ReadHoldingRegisters(block.StartRegister, block.RegisterCount)
			if err != nil {
				if manager.FailOnError {
					return nil, err
				}
				log.WithError(err).Warning("ignoring error on read holding registers (failOnError is false)")
				continue
			}

			// Trim the result bytes to the expected length of bytes as per the calculated
			// register count. This ensures there are no padded values included in the data
			// for subsequent processing.
			if len(results) > 0 {
				results = results[0 : 2*block.RegisterCount] // Two bytes per register
			}
			log.WithField("results", results).Debug("got block read results for holding registers")
			block.Results = results

			// Parse the results from the bulk read. This will create the readings
			// for each device.
			for _, device := range block.Devices {
				out := output.Get(device.Device.Output)

				reading, err := UnpackRegisterReading(out, block, device)
				if err != nil {
					if manager.FailOnError {
						return nil, err
					}
					log.WithError(err).Warning("ignoring error on register unpack (failOnError is false)")
					continue
				}

				log.WithFields(log.Fields{
					"device": device.Device.GetID(),
					"type":   reading.Type,
					"value":  reading.Value,
				}).Debug("created new reading from holding register")
				readCtx := sdk.NewReadContext(device.Device, []*output.Reading{reading})
				readings = append(readings, readCtx)
			}
		}
	}

	return readings, nil
}

// writeHoldingRegister validates and writes the provided data to the specified modbus register
// for the given modbus client.
//
// This is broken apart from the device handler write function to make it easier to test.
func writeHoldingRegister(client modbus.Client, register uint16, data *sdk.WriteData) error {
	// Translate the configured holding register data into a format accepted
	// by the modbus client.
	registerData, err := getHoldingRegisterData(data.Data)
	if err != nil {
		log.WithError(err).Error("failed to parse holding register write data")
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
