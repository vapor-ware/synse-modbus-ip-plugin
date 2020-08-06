package devices

import (
	"fmt"

	"github.com/goburrow/modbus"
	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
	"github.com/vapor-ware/synse-sdk/sdk"
)

// CoilsHandler is a handler that should be used for all devices/outputs
// that read from/write to coils.
var CoilsHandler = sdk.DeviceHandler{
	Name:     "coil",
	BulkRead: bulkReadCoils,
	Write:    writeCoils,
}

// TODO: Read only coils.

// bulkReadCoils performs a bulk read on the devices parameter reducing round trips.
func bulkReadCoils(devices []*sdk.Device) (readContexts []*sdk.ReadContext, err error) {

	log.Debugf("----------- bulkReadCoils start ---------------")

	// Ideally this would be done in setup, but for now this should work.
	// Map out the bulk read.
	bulkReadMap, keyOrder, err := MapBulkRead(devices, true)
	if err != nil {
		return nil, err
	}
	log.Debugf("bulkReadMap: %#v", bulkReadMap)

	// Perform the bulk reads.
	for a := 0; a < len(keyOrder); a++ {
		k := keyOrder[a]
		v := bulkReadMap[k]
		log.Debugf("bulkReadMap[%#v]: %#v", k, v)
		//fmt.Printf("Coils: bulkReadMap[%#v]: %#v", k, v)

		// New connection for each key.
		var client modbus.Client
		var modbusDeviceData *config.ModbusDeviceData
		client, modbusDeviceData, err = GetBulkReadClient(k)
		if err != nil {
			return nil, err
		}

		// For read in v, perform each read (modbus network call).
		for i := 0; i < len(v); i++ {
			read := v[i]
			log.Debugf("Reading bulkReadMap[%#v][%#v]", k, read)

			var readResults []byte
			fmt.Printf("*** Coil read. StartRegister: %v, RegisterCount %v\n", read.StartRegister, read.RegisterCount)
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

	deviceData, client, err := GetModbusDeviceDataAndClient(device)
	if err != nil {
		return err
	}

	// Pull out the data to send on the wire from data.Data.
	modbusData := data.Data
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

	// Write the coil data to the requested address.
	log.Debugf("Writing coil 0x%x, data 0x%x", deviceData.Address, coilData)
	_, err = (*client).WriteSingleCoil(deviceData.Address, coilData)
	incrementModbusCallCounter()
	return err
}
