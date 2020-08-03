package devices

import (
	"fmt"

	"github.com/goburrow/modbus"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
	"github.com/vapor-ware/synse-sdk/sdk"
)

var sortOrdinalSetForCoils = false

// CoilsHandler is a handler that should be used for all devices/outputs
// that read from/write to coils.
var CoilsHandler = sdk.DeviceHandler{
	Name:     "coil",
	BulkRead: bulkReadCoils,
	Write:    writeCoils,
}

// bulkReadCoils performs a bulk read on the devices parameter reducing round trips.
func bulkReadCoils(devices []*sdk.Device) (readContexts []*sdk.ReadContext, err error) {

	log.Debugf("----------- bulkReadCoils start ---------------")

	// Ideally this would be done in setup, but for now this should work.
	// Map out the bulk read.
	bulkReadMap, keyOrder, err := MapBulkRead(devices, !sortOrdinalSetForCoils, true)
	if err != nil {
		return nil, err
	}
	log.Debugf("bulkReadMap: %#v", bulkReadMap)
	sortOrdinalSetForCoils = true

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
			fmt.Printf("*** MODBUS CALL (coils) ***\n")
			readResults, err = client.ReadCoils(read.StartRegister, read.RegisterCount)
			incrementModbusCallCounter()
			if err != nil {
				log.Errorf("modbus bulk read coils failure: %v", err.Error())
				if modbusDeviceData.FailOnError {
					return nil, err
				}
				// No data from device. If fail on error is false, we should keep trying the remaining reads.
				read.ReadResults = []byte{}
				continue
			}
			log.Debugf("ReadCoils: results: 0x%0x, len(results) 0x%0x", readResults, len(readResults))
			read.ReadResults = readResults[0 : 2*(read.RegisterCount)] // Store raw results. Two bytes per register. TODO: Double check this.
		} // end for each read
	} // end for each modbus connection

	readContexts, err = MapBulkReadData(bulkReadMap, keyOrder)
	return
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

	//output, err := output.GetOutput(device.Output)
	//output := output.Get(device.Output)
	//if err != nil {
	//	return err
	//}

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

	//register := (*output).Data["address"]
	//registerInt, ok := register.(int)
	//if !ok {
	//	return fmt.Errorf("Unable to convert (*output).Data[address] to uint16: %v", (*output).Data["address"])
	//}

	var deviceData config.ModbusDeviceData
	err = mapstructure.Decode(device.Data, &deviceData)
	if err != nil {
		return
	}

	//registerUint16 := uint16(registerInt)
	//registerUint16 := data.Address
	registerUint16 := deviceData.Address

	log.Debugf("Writing coil 0x%x, data 0x%x", registerUint16, coilData)
	_, err = (*client).WriteSingleCoil(registerUint16, coilData)
	return err
}

/*
	"errors"
	"fmt"
	"strings"

	"github.com/goburrow/modbus"
	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/output"
*/

/*
// Handler name definitions. These are used in the modbus device config
// (under the "handler" key) to relate devices to a handler.
const handlerCoil = "coil"
const handlerReadOnlyCoil = "read_only_coil"

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

// ReadOnlyCoilsHandler is a device handler used to read from and write to modbus coils.
var ReadOnlyCoilsHandler = sdk.DeviceHandler{
	Name: handlerReadOnlyCoil,
	BulkRead: func(devices []*sdk.Device) (contexts []*sdk.ReadContext, e error) {
		managers, found := DeviceManagers[handlerCoil]
		if !found {
			return nil, errors.New("no device manager(s) found for coil handler")
		}
		return bulkReadCoils(managers)
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
		log.WithError(err).Error("[modbus] failed to parse coil write data")
		return err
	}

	log.WithFields(log.Fields{
		"address": register,
		"data":    coilData,
	}).Debug("[modbus] writing to coil")
	_, err = client.WriteSingleCoil(register, coilData)
	return err
}

// bulkReadCoils performs a bulk read for modbus devices configured to use the coils handler.
// It gets those devices from the device manager(s) associated with the handler, which is
// populated on plugin startup.
func bulkReadCoils(managers []*ModbusDeviceManager) ([]*sdk.ReadContext, error) {
	var readings []*sdk.ReadContext
	var managersInErr []*ModbusDeviceManager

	for _, manager := range managers {
		log.WithFields(log.Fields{
			"host":     manager.Host,
			"port":     manager.Port,
			"slave id": manager.SlaveID,
			"address":  manager.Address,
		}).Debug("[modbus] starting bulk read for coils")

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
			}).Debug("[modbus] reading coils for block")

			fmt.Printf("Calling modbus.ReadCoils: StartRegister %v, RegisterCount: %v\n", block.StartRegister, block.RegisterCount)
			results, err := client.ReadCoils(block.StartRegister, block.RegisterCount)
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
					}).Warn("[modbus] failed to read coils - skipping configured host")
					break
				}
				log.WithError(err).Warning("[modbus] ignoring error on read coils (failOnError is false)")
				continue
			}

			// Trim the result bytes to the expected length of bytes as per the calculated
			// register count. This ensures there are no padded values included in the data
			// for subsequent processing.
			if len(results) > 0 {
				results = results[0 : 2*block.RegisterCount] // Two bytes per register TODO double check this
			}
			log.WithField("results", results).Debug("[modbus] got block read results for coils")
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
					log.WithError(err).Warning("[modbus] ignoring error on register unpack (failOnError is false)")
					continue
				}

				log.WithFields(log.Fields{
					"device": device.Device.GetID(),
					"type":   reading.Type,
					"value":  reading.Value,
				}).Debug("[modbus] created new reading from input register")
				readCtx := sdk.NewReadContext(device.Device, []*output.Reading{reading})
				readings = append(readings, readCtx)
			}
		}
	}

	if len(managersInErr) == len(managers) {
		return nil, errors.New("failed to read coils from all configured hosts")
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
*/
