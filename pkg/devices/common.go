package devices

// This file contains common modbus device code.
import (
	"fmt"
	"sort"

	log "github.com/Sirupsen/logrus"
	"github.com/goburrow/modbus"
	"github.com/mitchellh/mapstructure"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/utils"
	"github.com/vapor-ware/synse-sdk/sdk"
)

// NextSortOrdinal doles out the next sort ordinal for scan sort order.
var NextSortOrdinal = int32(1)

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

// GetOutput is a helper to get the first output for a device. Called by Write functions.
func GetOutput(device *sdk.Device) (output *sdk.Output, err error) {

	if device == nil {
		return nil, fmt.Errorf("device is nil")
	}
	if device.Outputs == nil {
		return nil, fmt.Errorf("device.Outputs is nil")
	}

	// Checking for one output for now. We need the register to write.
	// We could support multiple outputs in the future by matching against
	// info if we like.
	length := len(device.Outputs)
	if length == 0 {
		return nil, fmt.Errorf("no device outputs")
	}

	if length > 1 {
		return nil, fmt.Errorf("multiple device outputs unsupported")
	}
	return device.Outputs[0], nil
}

// UnpackCoilReading gets a coil (true / false) from a ReadCoils result buffer.
func UnpackCoilReading(output *sdk.Output, rawReading []byte, startAddress uint16, coilAddress uint16) (reading *sdk.Reading, err error) {
	log.Debugf("Unpack Coil. rawReading %x, startAddress: %v, coilAddress: %v", rawReading, startAddress, coilAddress)
	bitIndex := coilAddress - startAddress
	byteIndex := bitIndex / 8
	bitIndex = bitIndex % 8

	if int(byteIndex) >= len(rawReading) {
		return nil, fmt.Errorf("failed to get coil data")
	}

	coilByte := rawReading[byteIndex]
	mask := byte(0x01 << bitIndex)
	coilState := (coilByte & mask) != 0

	return output.MakeReading(coilState)
	// In this case we will not check the 'failOnError' flag because
	// this isn't an issue with reading the device, its a configuration
	// issue and the error should be noted.
}

// UnpackReading is a wrapper for CastToType and MakeReading.
func UnpackReading(output *sdk.Output, typeName string, rawReading []byte, failOnErr bool) (reading *sdk.Reading, err error) {

	// Cast the raw reading value to the specified output type
	data, err := utils.CastToType(typeName, rawReading)
	if err != nil {
		log.Errorf("Failed to case typeName: %v, rawReading: %x", typeName, rawReading)
		if failOnErr {
			return nil, err
		}
		return nil, nil // No reading.
	}

	reading, err = output.MakeReading(data)
	if err != nil {
		// In this case we will not check the 'failOnError' flag because
		// this isn't an issue with reading the device, its a configuration
		// issue and the error should be noted.
		return nil, err
	}
	return
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

// MaximumRegisterCount is the max number of registers to read in one modbus
// call. We may need to tune this for some devices (not clear). Technical max
// is 123 for ReadHoldingRegisters over IP.
const MaximumRegisterCount uint16 = 123

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
func NewModbusBulkRead(device *sdk.Device, startRegister uint16, registerCount uint16, isCoil bool) (read *ModbusBulkRead, err error) {
	if device == nil {
		return nil, fmt.Errorf("no device pointer given")
	}
	read = &ModbusBulkRead{
		StartRegister: startRegister,
		RegisterCount: registerCount,
		IsCoil:        isCoil,
	}
	read.Devices = append(read.Devices, device)
	log.Errorf("NewModbusBulkRead returning: %#v", read)
	return
}

// SortDevicesByRegister sorts the device list by modbus register.
// Used for bulk reads.
// Returns sortedRegisters which is a slice of uint16 register addresses in ascending order.
// Returns deviceMap which is a map of register to sdk.Device.
func SortDevicesByRegister(devices []*sdk.Device, setSortOrdinal bool) (sortedRegisters []uint16, deviceMap map[uint16]*sdk.Device, err error) {

	if devices == nil {
		return nil, nil, nil // Nothing to sort. Could arguably fail here.
	}
	deviceMap = make(map[uint16]*sdk.Device)

	// For each device.
	for i := 0; i < len(devices); i++ {
		device := devices[i]

		// Deserialize the modbus configuration.
		var deviceData config.ModbusDeviceData
		err = mapstructure.Decode(device.Data, &deviceData)
		if err != nil {
			return nil, nil, err
		}

		// For each device output.
		outputs := device.Outputs
		for j := 0; j < len(outputs); j++ {

			// Get the output registers.
			output := outputs[j]
			var outputData config.ModbusOutputData
			err := mapstructure.Decode(output.Data, &outputData)
			if err != nil { // This is a configuration issue, so fail hard here.
				log.Errorf(
					"SortDevicesByRegister failed parsing output.Data device at:[%v], device: %#v, output at:[%v], output: %#v",
					i, device, j, output)
				return nil, nil, err
			}
			outputDataAddress := uint16(outputData.Address) // TODO: Can we have uint16 in the config struct?

			// Add to locals.
			sortedRegisters = append(sortedRegisters, outputDataAddress)
			deviceMap[outputDataAddress] = device
		} // end for each output
	} // end for each device

	// Sort / trace.
	sort.SliceStable(sortedRegisters, func(i, j int) bool { return sortedRegisters[i] < sortedRegisters[j] })
	// Add SortOrdinal to all devices.
	if setSortOrdinal {
		for k := 0; k < len(sortedRegisters); k++ {
			deviceMap[sortedRegisters[k]].SortOrdinal = NextSortOrdinal
			NextSortOrdinal++
		}
	}
	log.Debugf("sortedRegisters: %#v", sortedRegisters)
	log.Debugf("deviceMap: %#v", deviceMap)
	return
}

// MapBulkRead maps the physical modbus device / connection information for all
// modbus devices to a map of each modbus bulk read call required to get all
// register data configured for the device.
func MapBulkRead(devices []*sdk.Device, setSortOrdinal bool, isCoil bool) (bulkReadMap map[ModbusBulkReadKey][]*ModbusBulkRead, err error) {
	log.Debugf("MapBulkRead start. devices: %+v", devices)

	// Sort the devices.
	sortedRegisters, sortedDevices, err := SortDevicesByRegister(devices, setSortOrdinal)
	if err != nil {
		log.Errorf("failed to sort devices")
		return nil, err
	}
	bulkReadMap = make(map[ModbusBulkReadKey][]*ModbusBulkRead)

	for i := 0; i < len(sortedRegisters); i++ {
		// Create the key for this device from the device data.
		device := sortedDevices[sortedRegisters[i]]
		log.Debugf("--- next synse device: %v", device)
		var deviceData config.ModbusDeviceData
		err = mapstructure.Decode(device.Data, &deviceData)
		if err != nil {
			return nil, err
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

		// For each device output.
		outputs := device.Outputs
		for j := 0; j < len(outputs); j++ {
			output := outputs[j]
			var outputData config.ModbusOutputData
			// Get the output data. Need address and width.
			err := mapstructure.Decode(output.Data, &outputData)
			if err != nil { // Hard failure on configuration issue.
				log.Errorf(
					"MapBulkRead failed parsing output.Data device at:[%v], device: %#v, output at:[%v], output: %#v",
					i, device, j, output)
				return nil, err
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
					return nil, err
				}
				log.Debugf("modbusBulkRead: %#v", modbusBulkRead)
				bulkReadMap[key] = append(bulkReadMap[key], modbusBulkRead)
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
						return nil, err
					}
					log.Debugf("modbusBulkRead: %#v", modbusBulkRead)
					bulkReadMap[key] = append(bulkReadMap[key], modbusBulkRead)
				}
			}
		} // For each output
	} // For each device.
	return bulkReadMap, nil
}

// MapBulkReadData maps the data read over modbus to the device read contexts.
func MapBulkReadData(bulkReadMap map[ModbusBulkReadKey][]*ModbusBulkRead) (readContexts []*sdk.ReadContext, err error) {
	log.Debugf("MapBulkReadData start. bulkReadMap %+v", bulkReadMap)
	for k, v := range bulkReadMap {
		for h := 0; h < len(v); h++ { // for each read
			read := v[h]
			devices := read.Devices
			for i := 0; i < len(devices); i++ {
				device := devices[i]

				// For each device output.
				outputs := device.Outputs
				readings := []*sdk.Reading{}
				for j := 0; j < len(outputs); j++ {
					output := outputs[j]
					var outputData config.ModbusOutputData
					// Get the output data. Need address and width.
					err := mapstructure.Decode(output.Data, &outputData)
					if err != nil { // This is not a configuration issue. Device may not have responded.
						log.Errorf(
							"MapBulkReadData failed parsing output.Data device at:[%v], device: %#v, output at:[%v], output: %#v",
							i, device, j, output)
						if k.FailOnError {
							return nil, err
						}
					}
					outputDataAddress := uint16(outputData.Address) // TODO: Can we have uint16 in the config struct?
					outputDataWidth := uint16(outputData.Width)     // TODO: As above.

					log.Debugf("outputDataAddress: 0x%04x", outputDataAddress)
					log.Debugf("outputDataWidth: %d", outputDataWidth)
					log.Debugf("k.FailOnError: %v", k.FailOnError)

					readResults := read.ReadResults // Raw byte results from modbus call.

					var reading *sdk.Reading
					if read.IsCoil {
						reading, err = UnpackCoilReading(output, read.ReadResults, read.StartRegister, outputDataAddress)
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
							log.Errorf("Failed reading. Attempt to read beyond bounds. startDataOffset: %v, endDataOffset: %v, readResultsLength: %v",
								startDataOffset, endDataOffset, readResultsLength)
							readings = append(readings, nil)
							continue // Next output.
						} // end bounds check.

						rawReading := readResults[startDataOffset:endDataOffset]
						log.Debugf("rawReading: len: %v, %x", len(rawReading), rawReading)

						reading, err = UnpackReading(output, outputData.Type, rawReading, k.FailOnError)
						if err != nil {
							return nil, err
						}
					}
					log.Debugf("Appending reading: %#v", reading)
					readings = append(readings, reading)

				} // End for each output.
				readContext := sdk.NewReadContext(device, readings)
				readContexts = append(readContexts, readContext)
			} // End for each device.
		} // End for each read.
	} // End for each key, value.
	return
}
