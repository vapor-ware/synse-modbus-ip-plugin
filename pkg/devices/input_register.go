package devices

import (
	"errors"

	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/output"
)

const handlerInputRegister = "input_register"

// InputRegisterHandler is a handler that should be used for all devices/outputs
// that read input registers.
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

func bulkReadInputRegisters(managers []*ModbusDeviceManager) ([]*sdk.ReadContext, error) {
	var readings []*sdk.ReadContext

	for _, manager := range managers {
		err := manager.ParseBlocks()
		if err != nil {
			return nil, err
		}
		for _, block := range manager.Blocks {

			results, err := manager.Client.ReadInputRegisters(block.StartRegister, block.RegisterCount)
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

			// Parse the results from the bulk read. This will create the readings
			// for each device.
			for _, device := range block.Devices {
				out := output.Get(device.Device.Output)

				reading, err := UnpackRegisterReading(out, block.Results, block.StartRegister, device.Config.Address, device.Config.Width, device.Config.Type, device.Config.FailOnError)
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

//// bulkReadInputRegisters performs a bulk read on the devices parameter
//// reducing round trips to the physical device.
//func bulkReadInputRegistersOrig(devices []*sdk.Device) (readContexts []*sdk.ReadContext, err error) {
//	log.Errorf("----------- bulkReadInputRegisters start ---------------")
//
//	// Ideally this would be done in setup, but for now this should work.
//	// Map out the bulk read.
//	bulkReadMap, keyOrder, err := MapBulkRead(devices, !sortOrdinalSetForInput, false)
//	if err != nil {
//		return nil, err
//	}
//	log.Debugf("bulkReadMap: %#v", bulkReadMap)
//	sortOrdinalSetForInput = true
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
//			readResults, err = client.ReadInputRegisters(read.StartRegister, read.RegisterCount)
//			if err != nil {
//				log.Errorf("modbus bulk read input registers failure: %v", err.Error())
//				if modbusDeviceData.FailOnError {
//					return nil, err
//				}
//				// No data from device. If fail on error is false, we should keep trying the remaining reads.
//				read.ReadResults = []byte{}
//				continue
//			}
//			log.Debugf("ReadInputRegisters: results: 0x%0x, len(results) 0x%0x", readResults, len(readResults))
//			read.ReadResults = readResults[0 : 2*(read.RegisterCount)] // Store raw results. Two bytes per register.
//		} // end for each read
//	} // end for each modbus connection
//
//	readContexts, err = MapBulkReadData(bulkReadMap, keyOrder)
//	return
//}
