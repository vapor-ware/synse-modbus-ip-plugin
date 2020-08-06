package devices

// This file contains common modbus device code.
import (
	"fmt"
	"math"
	"sort"
	"sync"

	"github.com/goburrow/modbus"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/utils"
	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/output"
)

// TODO: We need to use MaximumCoilCount.
// TODO: We need read only coil and holding register.

// MaximumRegisterCount is The technical maximum is 123 for ReadHoldingRegisters over IP.
const MaximumRegisterCount uint16 = 123

// MaximumCoilCount is MaximumRegisterCount * 16 because 8 coil reading per byte and a register is two bytes.
// TODO: We need to use this.
const MaximumCoilCount uint16 = MaximumRegisterCount * 16

// GetModbusClientAndConfig is common code to get the modbus configuration and client from the device configuration.
func GetModbusClientAndConfig(device *sdk.Device) (modbusConfig *config.ModbusDeviceData, client *modbus.Client, err error) {

	// Pull the modbus configuration out of the device Data fields.
	var deviceData config.ModbusDeviceData
	err = mapstructure.Decode(device.Data, &deviceData)
	if err != nil {
		return nil, nil, err
	}

	// Create the modbus client from the configuration data.
	cli, err := utils.NewClient(&deviceData)
	if err != nil {
		return nil, nil, err
	}
	return &deviceData, &cli, nil
}

// GetBulkReadClient gets the modbus client and device data for the
// connection information in k.
func GetBulkReadClient(k ModbusBulkReadKey) (client modbus.Client, modbusDeviceData *config.ModbusDeviceData, err error) {
	log.Debugf("Creating modbus connection")
	modbusDeviceData = &config.ModbusDeviceData{
		Host:        k.Host,
		Port:        k.Port,
		Timeout:     k.Timeout,
		FailOnError: k.FailOnError,
		// Omitting SlaveID for now. Not currently used.
	}
	log.Debugf("modbusDeviceData: %#v", modbusDeviceData)
	client, err = utils.NewClient(modbusDeviceData)
	if err != nil {
		log.Debugf("modbus NewClient failure: %v", err.Error())
		if modbusDeviceData.FailOnError {
			return nil, nil, err
		}
	}
	return
}

// UnpackCoilReading gets a coil (true / false) from a ReadCoils result buffer.
func UnpackCoilReading(output *output.Output, rawReading []byte, startAddress uint16, coilAddress uint16, failOnErr bool) (
	reading *output.Reading, err error) {
	log.Debugf("Unpack Coil. rawReading %x, startAddress: %v, coilAddress: %v", rawReading, startAddress, coilAddress)
	bitIndex := coilAddress - startAddress
	byteIndex := bitIndex / 8
	bitIndex = bitIndex % 8

	if int(byteIndex) >= len(rawReading) {
		if failOnErr {
			return nil, fmt.Errorf("failed to get coil data")
		}
		return nil, nil // No Reading
	}

	coilByte := rawReading[byteIndex]
	mask := byte(0x01 << bitIndex)
	coilState := (coilByte & mask) != 0

	return output.MakeReading(coilState), nil
	// In this case we will not check the 'failOnError' flag because
	// this isn't an issue with reading the device, its a configuration
	// issue and the error should be noted.
}

// UnpackReading is a wrapper for CastToType and MakeReading.
func UnpackReading(output *output.Output, typeName string, rawReading []byte, failOnErr bool) (reading *output.Reading, err error) {

	// Cast the raw reading value to the specified output type
	data, err := utils.CastToType(typeName, rawReading)
	if err != nil {
		log.Errorf("Failed to cast typeName: %v, rawReading: %x", typeName, rawReading)
		if failOnErr {
			return nil, err
		}
		return nil, nil // No reading.
	}

	//fmt.Printf("*** output: %T, %#v\n", output, output)
	//fmt.Printf("*** data: %T, %#v\n", data, data)
	return output.MakeReading(data), nil
}

// ModbusBulkReadKey corresponds to a Modbus Device / Connection.
// We will need one or more bulk reads per key entry.
type ModbusBulkReadKey struct {
	// Modbus device host name.
	Host string
	// Modbus device port.
	Port int
	// Timeout for modbus read.
	Timeout string
	// Fail on error. (Do we abort on one failed read?)
	FailOnError bool
	// Maximum number of registers to read on a single modbus call to the device.
	MaximumRegisterCount uint16
}

// NewModbusBulkReadKey creates a modbus bulk read key.
func NewModbusBulkReadKey(host string, port int, timeout string, failOnError bool) (key *ModbusBulkReadKey, err error) {
	if host == "" {
		return nil, fmt.Errorf("empty host")
	}
	if port <= 0 {
		return nil, fmt.Errorf("invalid port %v", port)
	}
	key = &ModbusBulkReadKey{
		Host:                 host,
		Port:                 port,
		Timeout:              timeout,
		FailOnError:          failOnError,
		MaximumRegisterCount: MaximumRegisterCount,
	}
	return
}

// ModbusBulkRead contains data for each individual bulk read call to the modbus device.
type ModbusBulkRead struct {
	// Synse devices associated with this read.
	Devices []*sdk.Device
	// Raw Modbus read results
	ReadResults []byte
	// First register to read.
	StartRegister uint16
	// Number of registers to read.
	RegisterCount uint16
	// true for coils. The unmarshalling is different.
	IsCoil bool
}

// NewModbusBulkRead contains data for each bulk read.
func NewModbusBulkRead(device *sdk.Device, startRegister uint16, registerCount uint16, isCoil bool) (
	read *ModbusBulkRead, err error) {
	if device == nil {
		return nil, fmt.Errorf("no device pointer given")
	}
	read = &ModbusBulkRead{
		StartRegister: startRegister,
		RegisterCount: registerCount,
		IsCoil:        isCoil,
	}
	read.Devices = append(read.Devices, device)
	log.Debugf("NewModbusBulkRead returning: %#v", read)
	return
}

// ModbusDevice is an intermediate struct for sorting ModbusBulkReadKey.
type ModbusDevice struct {
	Host     string
	Port     int
	Register uint16
}

// SortDevices sorts the device list.
// Used for bulk reads.
// Returns sorted which is a slice of ModbusDevice in ascending order.
// Returns deviceMap which is a map of register to sdk.Device.
func SortDevices(devices []*sdk.Device, setSortOrdinal bool) (
	sorted []ModbusDevice, deviceMap map[ModbusDevice]*sdk.Device, err error) {

	if devices == nil {
		return nil, nil, nil // Nothing to sort. Could arguably fail here.
	}
	deviceMap = make(map[ModbusDevice]*sdk.Device)

	// For each device.
	for i := 0; i < len(devices); i++ {
		device := devices[i]

		// Deserialize the modbus configuration.
		var deviceData config.ModbusDeviceData
		err = mapstructure.Decode(device.Data, &deviceData)
		if err != nil {
			return nil, nil, err
		}

		key := ModbusDevice{
			Host:     deviceData.Host,
			Port:     deviceData.Port,
			Register: deviceData.Address, // TODO: Can we have uint16 in the config struct.
		}

		// Add to locals.
		sorted = append(sorted, key)
		deviceMap[key] = device
	} // end for each device

	// Sort / trace.
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].Host < sorted[j].Host {
			return true
		} else if sorted[i].Host > sorted[j].Host {
			return false
		}
		if sorted[i].Port < sorted[j].Port {
			return true
		} else if sorted[i].Port > sorted[j].Port {
			return false
		}
		if sorted[i].Register < sorted[j].Register {
			return true
		} else if sorted[i].Register > sorted[j].Register {
			return false
		}
		log.Errorf("Duplicate modbus device configured. Host: %v, Port: %v, Register: %v",
			sorted[i].Host, sorted[i].Port, sorted[i].Register)
		return true
	})

	return
}

// MapBulkRead maps the physical modbus device / connection information for all
// modbus devices to a map of each modbus bulk read call required to get all
// register data configured for the device.
func MapBulkRead(devices []*sdk.Device, setSortOrdinal bool, isCoil bool) (
	bulkReadMap map[ModbusBulkReadKey][]*ModbusBulkRead, keyOrder []ModbusBulkReadKey, err error) {
	log.Debugf("MapBulkRead start. devices: %+v", devices)
	for z := 0; z < len(devices); z++ {
		log.Debugf("MapBulkRead devices[%v]: %#v", z, devices[z])
		//fmt.Printf("MapBulkRead devices[%v]: %#v\n", z, devices[z])
	}

	// Sort the devices.
	sorted, sortedDevices, err := SortDevices(devices, setSortOrdinal)
	if err != nil {
		log.Errorf("failed to sort devices")
		return nil, keyOrder, err
	}
	bulkReadMap = make(map[ModbusBulkReadKey][]*ModbusBulkRead)

	for z := 0; z < len(sorted); z++ {
		log.Debugf("MapBulkRead sorted[%v]: %#v", z, sorted[z])
		//fmt.Printf("MapBulkRead sorted[%v]: %#v\n", z, sorted[z])
	}

	for i := 0; i < len(sorted); i++ {
		// Create the key for this device from the device data.
		device := sortedDevices[sorted[i]]
		log.Debugf("--- next synse device: %v", device)
		var deviceData config.ModbusDeviceData
		err = mapstructure.Decode(device.Data, &deviceData)
		if err != nil {
			return nil, keyOrder, err
		}

		key := ModbusBulkReadKey{
			Host:                 deviceData.Host,
			Port:                 deviceData.Port,
			Timeout:              deviceData.Timeout,
			FailOnError:          deviceData.FailOnError,
			MaximumRegisterCount: MaximumRegisterCount,
		}
		log.Debugf("Created key: %#v", key)

		// Find out if the key is in the map.
		keyValues, keyPresent := bulkReadMap[key]
		if keyPresent {
			log.Debugf("key is already in the map")
		} else {
			log.Debugf("key is not in the map")
		}

		log.Debugf("len(keyValues): %v", len(keyValues))

		var outputData config.ModbusDeviceData // TODO: Change outputData to deviceData
		err := mapstructure.Decode(device.Data, &outputData)
		if err != nil { // Hard failure on configuration issue.
			log.Errorf(
				//"MapBulkRead failed parsing output.Data device at:[%v], device: %#v, output at:[%v], output: %#v",
				"MapBulkRead failed parsing output.Data device at:[%v], device: %#v",
				//i, device, j, output)
				i, device)
			return nil, keyOrder, err
		}

		outputDataAddress := uint16(outputData.Address) // TODO: Can we have uint16 in the config struct?
		outputDataWidth := uint16(outputData.Width)     // TODO: As above.

		log.Debugf("outputDataAddress: 0x%04x", outputDataAddress)
		log.Debugf("outputDataWidth: %d", outputDataWidth)

		// Insert.
		// If the key is not present, this is a simple insert to the map.
		if !keyPresent {
			log.Debugf("Key not present.")
			modbusBulkRead, err := NewModbusBulkRead(device, outputDataAddress, outputDataWidth, isCoil)
			if err != nil {
				return nil, keyOrder, err
			}
			log.Debugf("modbusBulkRead: %#v", modbusBulkRead)
			bulkReadMap[key] = append(bulkReadMap[key], modbusBulkRead)
			keyOrder = append(keyOrder, key)
		} else {
			log.Debugf("Key present")
			// See if this fits on the end of the slice.
			//  If so, update the ModbusBulkRead RegisterCount.
			//  If not, create a new ModbusBulkRead.
			reads := bulkReadMap[key]
			lastRead := reads[len(reads)-1]
			startRegister := lastRead.StartRegister
			log.Debugf("startRegister: 0x%0x", startRegister)
			newRegisterCount := outputDataAddress + outputDataWidth - startRegister

			if newRegisterCount < key.MaximumRegisterCount {
				log.Debugf("read fits in existing. newRegisterCount: %v", newRegisterCount)
				lastRead.RegisterCount = newRegisterCount
				lastRead.Devices = append(lastRead.Devices, device)
			} else {
				// Add a new read.
				log.Debugf("read does not fit in existing. newRegisterCount: %v", newRegisterCount)
				modbusBulkRead, err := NewModbusBulkRead(device, outputDataAddress, outputDataWidth, isCoil)
				if err != nil {
					return nil, keyOrder, err
				}
				log.Debugf("modbusBulkRead: %#v", modbusBulkRead)
				bulkReadMap[key] = append(bulkReadMap[key], modbusBulkRead)
				keyOrder = append(keyOrder, key)
			}
		}
		//} // For each output
	} // For each device.
	return bulkReadMap, keyOrder, nil
}

// MapBulkReadData maps the data read over modbus to the device read contexts.
func MapBulkReadData(bulkReadMap map[ModbusBulkReadKey][]*ModbusBulkRead, keyOrder []ModbusBulkReadKey) (
	readContexts []*sdk.ReadContext, err error) {
	// This map tells us if we have already created a read context for this
	// device and output. We can hit the same device and output more than once in
	// this loop when there are multiple modbus reads for a single device (more
	// than 123 register addresses)
	accountedFor := make(map[*sdk.Device]*output.Output)

	for a := 0; a < len(keyOrder); a++ {
		k := keyOrder[a]
		v := bulkReadMap[k]
		for h := 0; h < len(v); h++ { // for each read
			read := v[h]
			devices := read.Devices

			for i := 0; i < len(devices); i++ {
				device := devices[i]

				theOutput := output.Get(device.Output)
				readings := []*output.Reading{}

				// Have we accounted for this device and output yet?
				// This can happen when multiple reads are required for a single ModbusBulkReadKey.
				_, keyPresent := accountedFor[device]
				inMap := false
				if keyPresent {
					// Device is there. Is the output there?
					if accountedFor[device] == theOutput {
						inMap = true
						break // for
					}
					//}
					if inMap {
						log.Debugf("device[output] already accounted for: device %p, output %p", device, theOutput)
						continue // next output
					}
				}

				// Get address and width.
				var deviceData config.ModbusDeviceData
				err := mapstructure.Decode(device.Data, &deviceData)
				if err != nil { // This is not a configuration issue. Device may not have responded.
					log.Errorf(
						//"MapBulkReadData failed parsing output.Data device at:[%v], device: %#v, output at:[%v], output: %#v",
						//i, device, j, output)
						"MapBulkReadData failed parsing device at:[%v], device: %#v",
						i, device)
					if k.FailOnError {
						return nil, err
					}
				}

				// TODO: Variable names / Trace names.
				outputDataAddress := uint16(deviceData.Address) // TODO: Can we have uint16 in the config struct?
				outputDataWidth := uint16(deviceData.Width)     // TODO: As above.

				log.Debugf("outputDataAddress: 0x%04x", outputDataAddress)
				log.Debugf("outputDataWidth: %d", outputDataWidth)
				log.Debugf("k.FailOnError: %v", k.FailOnError)

				readResults := read.ReadResults // Raw byte results from modbus call.

				//var reading *sdk.Reading
				var reading *output.Reading
				if read.IsCoil {
					reading, err = UnpackCoilReading(theOutput, read.ReadResults, read.StartRegister, outputDataAddress, k.FailOnError)
					if err != nil {
						return nil, err
					}
				} else {
					// Get start and end data offsets. Bounds check.
					startDataOffset := (2 * outputDataAddress) - (2 * read.StartRegister) // Results are in bytes. Need 16 bit words.
					endDataOffset := startDataOffset + (2 * outputDataWidth)              // Two bytes per register.
					readResultsLength := len(readResults)

					log.Debugf("startDataOffset: %d", startDataOffset)
					log.Debugf("endDataOffset: %d", endDataOffset)
					log.Debugf("readResultsLength: %d", readResultsLength)

					if int(endDataOffset) > len(readResults) {
						if k.FailOnError {
							return nil, fmt.Errorf("Bounds check failure. startDataOffset: %v, endDataOffset: %v, readResultsLength: %v",
								startDataOffset, endDataOffset, readResultsLength)
						}
						// nil reading.
						log.Errorf("No data. Attempt to read beyond bounds. startDataOffset: %v, endDataOffset: %v, readResultsLength: %v",
							startDataOffset, endDataOffset, readResultsLength)
						readings = append(readings, nil)
						continue // Next output.
					} // end bounds check.

					rawReading := readResults[startDataOffset:endDataOffset]
					log.Debugf("rawReading: len: %v, %x", len(rawReading), rawReading)

					reading, err = UnpackReading(theOutput, deviceData.Type, rawReading, k.FailOnError)
					if err != nil {
						return nil, err
					}
				}
				log.Debugf("Appending reading: %#v, device: %v, output: %#v", reading, device, theOutput)
				readings = append(readings, reading)

				// Add to accounted for.
				//accountedFor[device] = append(accountedFor[device], theOutput)
				accountedFor[device] = theOutput

				//} // End for each output.

				// Only append a read context if we have readings. (Including nil readings)
				if len(readings) > 0 {
					readContext := sdk.NewReadContext(device, readings)
					readContexts = append(readContexts, readContext)
					//log.Debugf("Appending readContext: %#v, device: %v, outputs: %#v", readContext, device, outputs)
					log.Debugf("Appending readContext: %#v, device: %v", readContext, device)
				} else {
					log.Debugf("No readings to append. Not creating read context")
				}
			} // End for each device.
		} // End for each read.
	} // End for each key, value.
	return
}

// ModbusCallCounter is for testing. Here we increment it once per network round
// trip with any modbus server.
var modbusCallCounter uint64
var mutex sync.Mutex

// GetModbusCallCounter gets the number of modbus calls to any modbus server.
func GetModbusCallCounter() (counter uint64) {
	mutex.Lock()
	counter = modbusCallCounter
	mutex.Unlock()
	return
}

// ResetModbusCallCounter resets the counter to zero for test purposes.
func ResetModbusCallCounter() {
	mutex.Lock()
	modbusCallCounter = 0
	mutex.Unlock()
}

// incrementModbusCallCounter is called internally whenever a modbus request is
// made to any modbus server.
func incrementModbusCallCounter() {
	mutex.Lock()
	if modbusCallCounter == math.MaxUint64 {
		modbusCallCounter = 0
	} else {
		modbusCallCounter++
	}
	mutex.Unlock()
}
