package devices

import (
	"github.com/goburrow/modbus"
	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
	"github.com/vapor-ware/synse-sdk/sdk"
)

var sortOrdinalSetForInput = false

// InputRegisterHandler is a handler that should be used for all devices/outputs
// that read input registers.
var InputRegisterHandler = sdk.DeviceHandler{
	Name:     "input_register",
	BulkRead: bulkReadInputRegisters,
}

// bulkReadInputRegisters performs a bulk read on the devices parameter
// reducing round trips to the physical device.
func bulkReadInputRegisters(devices []*sdk.Device) (readContexts []*sdk.ReadContext, err error) {
	log.Errorf("----------- bulkReadInputRegisters start ---------------")

	// Ideally this would be done in setup, but for now this should work.
	// Map out the bulk read.
	bulkReadMap, keyOrder, err := MapBulkRead(devices, !sortOrdinalSetForInput, false)
	if err != nil {
		return nil, err
	}
	log.Debugf("bulkReadMap: %#v", bulkReadMap)
	sortOrdinalSetForInput = true

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
			readResults, err = client.ReadInputRegisters(read.StartRegister, read.RegisterCount)
			incrementModbusCallCounter()
			if err != nil {
				log.Errorf("modbus bulk read input registers failure: %v", err.Error())
				if modbusDeviceData.FailOnError {
					return nil, err
				}
				// No data from device. If fail on error is false, we should keep trying the remaining reads.
				read.ReadResults = []byte{}
				continue
			}
			log.Debugf("ReadInputRegisters: results: 0x%0x, len(results) 0x%0x", readResults, len(readResults))
			read.ReadResults = readResults[0 : 2*(read.RegisterCount)] // Store raw results. Two bytes per register.
		} // end for each read
	} // end for each modbus connection

	readContexts, err = MapBulkReadData(bulkReadMap, keyOrder)
	return
}

/*
	"errors"
	"fmt"

	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/output"
*/

/*
const handlerInputRegister = "input_register"

// InputRegisterHandler is a device handler used to read from modbus input registers.
var InputRegisterHandler = sdk.DeviceHandler{
	Name: handlerInputRegister,
	BulkRead: func(devices []*sdk.Device) (contexts []*sdk.ReadContext, e error) {
		managers, found := DeviceManagers[handlerInputRegister]
		if !found {
			return nil, errors.New("no device manager(s) found for input register handler")
		}
		return bulkReadInputRegisters(managers)
	},
}

// bulkReadInputRegisters performs a bulk read for modbus devices configured to use the input
// register handler. It gets those devices from the device manager(s) associated with the handler,
// which is populated on plugin startup.
func bulkReadInputRegisters(managers []*ModbusDeviceManager) ([]*sdk.ReadContext, error) {
	var readings []*sdk.ReadContext
	var managersInErr []*ModbusDeviceManager

	for _, manager := range managers {
		log.WithFields(log.Fields{
			"host":     manager.Host,
			"port":     manager.Port,
			"slave id": manager.SlaveID,
			"address":  manager.Address,
		}).Debug("[modbus] starting bulk read for input registers")

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
			}).Debug("[modbus] reading input registers for block")

			fmt.Printf("Calling modbus ReadInputRegisters\n")
			results, err := client.ReadInputRegisters(block.StartRegister, block.RegisterCount)
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
					}).Warn("[modbus] failed to read input registers - skipping configured host")
					break
				}
				log.WithError(err).Warning("[modbus] ignoring error on read input registers (failOnError is false)")
				continue
			}

			// Trim the result bytes to the expected length of bytes as per the calculated
			// register count. This ensures there are no padded values included in the data
			// for subsequent processing.
			if len(results) > 0 {
				results = results[0 : 2*block.RegisterCount] // Two bytes per register
			}
			log.WithField("results", results).Debug("[modbus] got block read results for input registers")
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
				}).Debug("[modbus] created new reading from input register")
				readCtx := sdk.NewReadContext(device.Device, []*output.Reading{reading})
				readings = append(readings, readCtx)
			}
		}
	}

	if len(managersInErr) == len(managers) {
		return nil, errors.New("failed to read input registers from all configured hosts")
	}
	return readings, nil
}
*/
