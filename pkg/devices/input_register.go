package devices

import (
	"github.com/goburrow/modbus"
	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
	"github.com/vapor-ware/synse-sdk/sdk"
)

// InputRegisterHandler is a handler that should be used for all devices/outputs
// that read input registers.
var InputRegisterHandler = sdk.DeviceHandler{
	Name:     "input_register",
	BulkRead: bulkReadInputRegisters,
}

// bulkReadInputRegisters performs a bulk read on the devices parameter
// reducing round trips to the physical device.
func bulkReadInputRegisters(devices []*sdk.Device) (readContexts []*sdk.ReadContext, err error) {
	log.Debugf("----------- bulkReadInputRegisters start ---------------")

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
		var deviceData *config.ModbusDeviceData
		client, deviceData, err = GetBulkReadClient(k)
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
			log.Debugf("[modbus call]: ReadInputRegisters(0x%x, 0x%x), result: %v, len(d%d), err: %v\n",
				read.StartRegister, read.RegisterCount, readResults, len(readResults), err)
			if err != nil {
				log.Errorf("modbus bulk read input registers failure: %v", err.Error())
				if deviceData.FailOnError {
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
