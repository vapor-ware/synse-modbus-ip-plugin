package devices

import (
	"fmt"
	"strconv"

	"github.com/goburrow/modbus"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
	"github.com/vapor-ware/synse-sdk/sdk"
)

var sortOrdinalSetForHolding = false

// HoldingRegisterHandler is a handler which should be used for all devices/outputs
// that read from/write to holding registers.
var HoldingRegisterHandler = sdk.DeviceHandler{
	Name:     "holding_register",
	BulkRead: bulkReadHoldingRegisters,
	Write:    writeHoldingRegister,
}

// bulkReadHoldingRegisters performs a bulk read on the devices parameter
// reducing round trips to the physical device.
func bulkReadHoldingRegisters(devices []*sdk.Device) (readContexts []*sdk.ReadContext, err error) {
	log.Debugf("----------- bulkReadHoldingRegisters start ---------------")

	// Ideally this would be done in setup, but for now this should work.
	// Map out the bulk read.
	bulkReadMap, keyOrder, err := MapBulkRead(devices, !sortOrdinalSetForHolding, false)
	if err != nil {
		return nil, err
	}
	log.Debugf("bulkReadMap: %#v", bulkReadMap)
	sortOrdinalSetForHolding = true

	// Perform the bulk reads.
	for a := 0; a < len(keyOrder); a++ {
		k := keyOrder[a]
		v := bulkReadMap[k]
		log.Debugf("bulkReadMap[%#v]: %#v", k, v)

		// New connection for each key.
		var client modbus.Client
		var modbusDeviceData *config.ModbusDeviceData
		client, modbusDeviceData, err = GetBulkReadClient(k)
		if err != nil {
			return nil, err
		}

		// For read in v, perform each read.
		for i := 0; i < len(v); i++ { // For each required read.
			read := v[i]
			log.Debugf("Reading bulkReadMap[%#v][%#v]", k, read)

			var readResults []byte
			readResults, err = client.ReadHoldingRegisters(read.StartRegister, read.RegisterCount)
			incrementModbusCallCounter()
			if err != nil {
				log.Errorf("modbus bulk read holding registers failure: %v", err.Error())
				if modbusDeviceData.FailOnError {
					return nil, err
				}
				// No data from device. If fail on error is false, we should keep trying the remaining reads.
				read.ReadResults = []byte{}
				continue
			}
			log.Debugf("ReadHoldingRegisters: results: 0x%0x, len(results) 0x%0x", readResults, len(readResults))
			fmt.Printf("ReadHoldingRegisters: results: 0x%0x, len(results) 0x%0x", readResults, len(readResults))
			read.ReadResults = readResults[0 : 2*(read.RegisterCount)] // Store raw results. Two bytes per register.
		} // end for each read
	} // end for each modbus connection

	readContexts, err = MapBulkReadData(bulkReadMap, keyOrder)
	return
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

	//output, err := GetOutput(device)
	//if err != nil {
	//	return err
	//}

	// Translate the data. This is currently a hex string.
	dataString := string(modbusData)
	register64, err := strconv.ParseUint(dataString, 16, 16)
	if err != nil {
		return fmt.Errorf("Unable to parse uint16 %v", dataString)
	}
	registerData := uint16(register64)

	//register := (*output).Data["address"]
	//registerInt, ok := register.(int)
	//if !ok {
	//	return fmt.Errorf("Unable to convert (*output).Data[address] to uint16: %v", (*output).Data["address"])
	//}
	//registerUint16 := uint16(registerInt)

	var deviceData config.ModbusDeviceData
	err = mapstructure.Decode(device.Data, &deviceData)
	if err != nil {
		return
	}

	//registerUint16 := uint16(registerInt)
	registerUint16 := deviceData.Address

	log.Debugf("Writing holding register 0x%x, data 0x%x", registerUint16, registerData)
	_, err = (*client).WriteSingleRegister(registerUint16, registerData)
	incrementModbusCallCounter()
	return err
}

/*
	"fmt"
	"strconv"

	"github.com/goburrow/modbus"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/output"
*/

/*
// Handler name definitions. These are used in the modbus device config
// (under the "handler" key) to relate devices to a handler.
const handlerHoldingRegister = "holding_register"
const handlerReadOnlyHoldingRegister = "read_only_holding_register"

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

// ReadOnlyHoldingRegisterHandler is a device handler used to read modbus
// holding registers. No write method is available. This can be used to
// restrict write access to holding registers
var ReadOnlyHoldingRegisterHandler = sdk.DeviceHandler{
	Name: handlerReadOnlyHoldingRegister,
	BulkRead: func(devices []*sdk.Device) (contexts []*sdk.ReadContext, e error) {
		managers, found := DeviceManagers[handlerReadOnlyHoldingRegister]
		if !found {
			return nil, errors.New("no device manager(s) found for read only holding register handler")
		}
		return bulkReadHoldingRegisters(managers)
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

		client, err := manager.NewClient()
		if err != nil {
			log.WithError(err).Error("[modbus] failed to create new client for manager")
			if !manager.FailOnError {
				continue
			}
			return nil, err
		}

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

			fmt.Printf("Callng modbus ReadHoldingRegisters\n")
			results, err := client.ReadHoldingRegisters(block.StartRegister, block.RegisterCount)
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
*/
