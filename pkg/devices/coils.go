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

			// TODO: check if this is needed.
			//if len(results) > 0 {
			//	block.Results = results[0:2*block.RegisterCount]
			//}
			block.Results = results

			// Parse the results from the bulk read. This will create the readings for
			// each device.
			for _, device := range block.Devices {
				out := output.Get(device.Device.Output)

				// TODO: this may need to be updated a bit? I feel like it is peculiar that the register
				//   width is not factored into the calculation for unpacking the readings.
				// TODO: maybe move the failOnError check out of theUnpackCoilReading fn and into this fn?
				reading, err := UnpackCoilReading(out, block.Results, block.StartRegister, device.Config.Address, device.Config.FailOnError)
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

//// bulkReadCoils performs a bulk read on the devices parameter reducing round trips.
//func bulkReadCoilsOrig(devices []*sdk.Device) (readContexts []*sdk.ReadContext, err error) {
//
//	log.Debugf("----------- bulkReadCoils start ---------------")
//
//	// Ideally this would be done in setup, but for now this should work.
//	// Map out the bulk read.
//	bulkReadMap, keyOrder, err := MapBulkRead(devices, !sortOrdinalSetForCoils, true)
//	if err != nil {
//		return nil, err
//	}
//	log.Debugf("bulkReadMap: %#v", bulkReadMap)
//	sortOrdinalSetForCoils = true
//
//	// Perform the bulk reads.
//	for a := 0; a < len(keyOrder); a++ {
//		k := keyOrder[a]
//		v := bulkReadMap[k]
//		log.Debugf("bulkReadMap[%#v]: %#v", k, v)
//
//		// New connection for each key.
//		var client modbus.Client
//		var modbusDeviceData *config.ModbusConfig
//		client, modbusDeviceData, err = GetBulkReadClient(k)
//		if err != nil {
//			return nil, err
//		}
//
//		// For read in v, perform each read.
//		for i := 0; i < len(v); i++ { // For each required read.
//			read := v[i]
//			log.Debugf("Reading bulkReadMap[%#v][%#v]", k, read)
//
//			var readResults []byte
//			readResults, err = client.ReadCoils(read.StartRegister, read.RegisterCount)
//			if err != nil {
//				log.Errorf("modbus bulk read coils failure: %v", err.Error())
//				if modbusDeviceData.FailOnError {
//					return nil, err
//				}
//				// No data from device. If fail on error is false, we should keep trying the remaining reads.
//				read.ReadResults = []byte{}
//				continue
//			}
//			log.Debugf("ReadCoils: results: 0x%0x, len(results) 0x%0x", readResults, len(readResults))
//			read.ReadResults = readResults[0 : 2*(read.RegisterCount)] // Store raw results. Two bytes per register. TODO: Double check this.
//		} // end for each read
//	} // end for each modbus connection
//
//	readContexts, err = MapBulkReadData(bulkReadMap, keyOrder)
//	return
//}

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
