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

// CoilsHandler is a device handler used to read from and write to modbus coils.
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

// writeCoil validates and writes the provided data to the specified modbus register
// for the given modbus client.
//
// This is broken apart from the device handler write function to make it easier to test.
func writeCoil(client modbus.Client, register uint16, data *sdk.WriteData) error {
	// Translate the configured coil data into a format accepted
	// by the modbus client.
	coilData, err := getCoilData(data.Data)
	if err != nil {
		log.WithError(err).Error("failed to parse coil write data")
		return err
	}

	log.WithFields(log.Fields{
		"address": register,
		"data":    coilData,
	}).Debug("writing to coil")
	_, err = client.WriteSingleCoil(register, coilData)
	return err
}

// bulkReadCoils performs a bulk read for modbus devices configured to use the coils handler.
// It gets those devices from the device manager(s) associated with the handler, which is
// populated on plugin startup.
func bulkReadCoils(managers []*ModbusDeviceManager) ([]*sdk.ReadContext, error) {
	var readings []*sdk.ReadContext

	log.Debug("starting bulk read for coils")
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
			}).Debug("reading coils for block")
			results, err := manager.Client.ReadCoils(block.StartRegister, block.RegisterCount)
			if err != nil {
				if manager.FailOnError {
					return nil, err
				}
				log.WithError(err).Warning("ignoring error on read coils (failOnError is false)")
				continue
			}

			// Trim the result bytes to the expected length of bytes as per the calculated
			// register count. This ensures there are no padded values included in the data
			// for subsequent processing.
			if len(results) > 0 {
				results = results[0 : 2*block.RegisterCount] // Two bytes per register TODO double check this
			}
			log.WithField("results", results).Debug("got block read results for coils")
			block.Results = results

			// Parse the results from the bulk read. This will create the readings for
			// each device. Results are parsed by using a Device's address and register
			// width as indexes into the results byte slice.
			for _, device := range block.Devices {
				out := output.Get(device.Device.Output)

				reading, err := UnpackCoilReading(out, block, device)
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
				}).Debug("created new reading from input register")
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
