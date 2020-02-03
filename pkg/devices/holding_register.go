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
	var managersInErr []*ModbusDeviceManager

	for _, manager := range managers {
		log.WithFields(log.Fields{
			"host":     manager.Host,
			"port":     manager.Port,
			"slave id": manager.SlaveID,
			"address":  manager.Address,
		}).Debug("[modbus] starting bulk read for holding registers")

		// Attempt to parse the manager's devices into register blocks for bulk read
		// if they have not already been parsed.
		if err := manager.ParseBlocks(); err != nil {
			log.WithError(err).Error("[modbus] failed to parse devices into read blocks")
			return nil, err
		}

		for _, block := range manager.Blocks {
			log.WithFields(log.Fields{
				"startRegister": block.StartRegister,
				"registerCount": block.RegisterCount,
			}).Debug("[modbus] reading holding registers for block")
			results, err := manager.Client.ReadHoldingRegisters(block.StartRegister, block.RegisterCount)
			if err != nil {
				if manager.FailOnError {
					// Since there may be multiple managers (e.g. modbus sources) configured,
					// we don't want a failure to connect/read from one host to fail the read
					// on another host. As such, we will skip over the manager on a read error
					// and try the next one. If all managers fail to read, the bulk read will
					// return an error.
					managersInErr = append(managersInErr, manager)
					log.WithFields(log.Fields{
						"host": manager.Host,
						"port": manager.Port,
					}).Warn("[modbus] failed to read holding registers - skipping configured host")
					break
				}
				log.WithError(err).Warning("[modbus] ignoring error on read holding registers (failOnError is false)")
				continue
			}

			// Trim the result bytes to the expected length of bytes as per the calculated
			// register count. This ensures there are no padded values included in the data
			// for subsequent processing.
			if len(results) > 0 {
				results = results[0 : 2*block.RegisterCount] // Two bytes per register
			}
			log.WithField("results", results).Debug("[modbus] got block read results for holding registers")
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
					log.WithError(err).Warning("[modbus] ignoring error on register unpack (failOnError is false)")
					continue
				}

				log.WithFields(log.Fields{
					"device": device.Device.GetID(),
					"type":   reading.Type,
					"value":  reading.Value,
				}).Debug("[modbus] created new reading from holding register")
				readCtx := sdk.NewReadContext(device.Device, []*output.Reading{reading})
				readings = append(readings, readCtx)
			}
		}
	}

	if len(managersInErr) == len(managers) {
		for _, manager := range managersInErr {
			log.WithFields(log.Fields{
				"host": manager.Host,
				"port": manager.Port,
			}).Error("[modbus] failed to read holding registers from host")
		}
		return nil, errors.New("failed to read holding registers from all configured hosts")
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
		log.WithError(err).Error("[modbus] failed to parse holding register write data")
		return err
	}

	log.WithFields(log.Fields{
		"address": register,
		"data":    registerData,
	}).Debug("[modbus] writing to holding register")
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
