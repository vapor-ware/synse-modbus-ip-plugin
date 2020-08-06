package devices

import (
	"fmt"
	"strconv"

	"github.com/goburrow/modbus"
	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
	"github.com/vapor-ware/synse-sdk/sdk"
)

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
	bulkReadMap, keyOrder, err := MapBulkRead(devices, false)
	if err != nil {
		return nil, err
	}
	log.Debugf("bulkReadMap: %#v", bulkReadMap)

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
			//fmt.Printf("ReadHoldingRegisters: results: 0x%0x, len(results) 0x%0x", readResults, len(readResults))
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

	deviceData, client, err := GetModbusDeviceDataAndClient(device)
	if err != nil {
		return err
	}

	// Pull out the data to send on the wire from data.Data.
	modbusData := data.Data
	// Translate the data. This is currently a hex string.
	dataString := string(modbusData)
	register64, err := strconv.ParseUint(dataString, 16, 16)
	if err != nil {
		return fmt.Errorf("Unable to parse uint16 %v", dataString)
	}
	registerData := uint16(register64)

	// Modbus write.
	register := deviceData.Address
	log.Debugf("Writing holding register 0x%x, data 0x%x", register, registerData)
	_, err = (*client).WriteSingleRegister(register, registerData)
	incrementModbusCallCounter()
	return err
}
