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

/*
	"errors"
	"fmt"
	"sort"

	"github.com/goburrow/modbus"
	"github.com/mitchellh/mapstructure"
	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/utils"
	"github.com/vapor-ware/synse-sdk/sdk"
	"github.com/vapor-ware/synse-sdk/sdk/output"
*/

// MaximumRegisterCount is The technical maximum is 123 for ReadHoldingRegisters over IP.
const MaximumRegisterCount uint16 = 123

// MaximumCoilCount is MaximumRegisterCount * 16 because 8 coil reading per byte and a register is two bytes.
// TODO: We need to use this.
const MaximumCoilCount uint16 = MaximumRegisterCount * 16

//// NextSortOrdinal doles out the next sort ordinal for scan sort order.
//var NextSortOrdinal = int32(1)

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

/*
// GetOutput is a helper to get the first output for a device. Called by Write functions.
//func GetOutput(device *sdk.Device) (output *sdk.Output, err error) {
func GetOutput(device *sdk.Device) (output *output.Output, err error) {

	if device == nil {
		return nil, fmt.Errorf("device is nil")
	}
	//if device.Outputs == nil {
	if output == nil {
		return nil, fmt.Errorf("output is nil")
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
*/

// UnpackCoilReading gets a coil (true / false) from a ReadCoils result buffer.
//func UnpackCoilReading(output *sdk.Output, rawReading []byte, startAddress uint16, coilAddress uint16, failOnErr bool) (
//	reading *sdk.Reading, err error) {
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
//func UnpackReading(output *sdk.Output, typeName string, rawReading []byte, failOnErr bool) (reading *sdk.Reading, err error) {
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
	/*
		reading, err = output.MakeReading(data)
		if err != nil {
			// In this case we will not check the 'failOnError' flag because
			// this isn't an issue with reading the device, its a configuration
			// issue and the error should be noted.
			return nil, err
		}
		return
	*/
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

//// MaximumRegisterCount is the max number of registers to read in one modbus
//// call. We may need to tune this for some devices (not clear). Technical max
//// is 123 for ReadHoldingRegisters over IP.
//const MaximumRegisterCount uint16 = 123

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
// TODO: private?
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

		// For each device output.
		//outputs := device.Outputs
		//for j := 0; j < len(outputs); j++ {

		// Get the data to form the key.
		//output := outputs[j]
		//var outputData config.ModbusOutputData
		//err := mapstructure.Decode(output.Data, &outputData)
		//if err != nil { // This is a configuration issue, so fail hard here.
		//	log.Errorf(
		//		"SortDevices failed parsing output.Data device at:[%v], device: %#v, output at:[%v], output: %#v",
		//		i, device, j, output)
		//	return nil, nil, err
		//}

		key := ModbusDevice{
			Host: deviceData.Host,
			Port: deviceData.Port,
			//Register: uint16(outputData.Address), // TODO: Can we have uint16 in the config struct.
			Register: deviceData.Address, // TODO: Can we have uint16 in the config struct.
		}

		// Add to locals.
		sorted = append(sorted, key)
		deviceMap[key] = device
		//} // end for each output
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

	// TODO: DO WE NEED TO RE-ADD THIS? NO sort ordinal anymore ...

	/*
		// Add SortOrdinal to all devices.
		if setSortOrdinal {
			for k := 0; k < len(sorted); k++ {
				deviceMap[sorted[k]].SortOrdinal = NextSortOrdinal
				NextSortOrdinal++
			}
		}
	*/
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
		fmt.Printf("MapBulkRead devices[%v]: %#v\n", z, devices[z])
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
		fmt.Printf("MapBulkRead sorted[%v]: %#v\n", z, sorted[z])
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

		//// For each device output.
		//outputs := device.Outputs
		//for j := 0; j < len(outputs); j++ {
		//output := outputs[j]
		//var outputData config.ModbusOutputData
		// Get the output data. Need address and width.
		//err := mapstructure.Decode(output.Data, &outputData)

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
	//accountedFor := make(map[*sdk.Device][]*sdk.Output)
	//accountedFor := make(map[*sdk.Device][]*output.Output)
	accountedFor := make(map[*sdk.Device]*output.Output)
	//accountedFor := make(map[*sdk.Device][]string)
	//var accountedFor [*sdk.Device] //:= make(map[*sdk.Device][]*sdk.Output)

	for a := 0; a < len(keyOrder); a++ {
		k := keyOrder[a]
		v := bulkReadMap[k]
		for h := 0; h < len(v); h++ { // for each read
			read := v[h]
			devices := read.Devices

			for i := 0; i < len(devices); i++ {
				device := devices[i]

				// For each device output.
				//outputs := device.Outputs
				//readings := []*sdk.Reading{}
				//for j := 0; j < len(outputs); j++ {
				//output := outputs[j]
				//output := device.Output
				theOutput := output.Get(device.Output)
				readings := []*output.Reading{}

				// Have we accounted for this device and output yet?
				// This can happen when multiple reads are required for a single ModbusBulkReadKey.
				_, keyPresent := accountedFor[device]
				inMap := false
				if keyPresent {
					// Device is there. Is the output there?
					//for b := 0; b < len(accountedFor[device]); b++ {
					//if accountedFor[device][b] == output {
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

				//var outputData config.ModbusOutputData
				//// Get the output data. Need address and width.
				//err := mapstructure.Decode(output.Data, &outputData)
				//if err != nil { // This is not a configuration issue. Device may not have responded.
				//	log.Errorf(
				//		"MapBulkReadData failed parsing output.Data device at:[%v], device: %#v, output at:[%v], output: %#v",
				//		i, device, j, output)
				//	if k.FailOnError {
				//		return nil, err
				//	}
				//}
				//outputDataAddress := uint16(outputData.Address) // TODO: Can we have uint16 in the config struct?
				//outputDataWidth := uint16(outputData.Width)     // TODO: As above.

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

					//reading, err = UnpackReading(theOutput, outputData.Type, rawReading, k.FailOnError)
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

/*

// ErrDevicesNotSorted is an error which specifies that the plugin is unable
// to parse devices into read blocks correctly because the devices have not
// yet been sorted.
var ErrDevicesNotSorted = errors.New("devices not sorted; unable to parse bulk read blocks")

*/

/*
// ModbusDeviceAndData wraps an SDK Device and associates it with a ModbusDeviceData
//// ModbusDevice wraps an SDK Device and associates it with a ModbusConfig
// configuration parsed from the SDK Device's Data field.
type ModbusDeviceAndData struct {
	//Config *config.ModbusConfig
	Data   *config.ModbusDeviceData
	Device *sdk.Device
}

// NewModbusDeviceAndData creates a new instance of the ModbusDeviceAndData
// wrapper for the given SDK Device. This is here to only have to deserialize
// the generic device.Data field to ModbusDeviceData once.
func NewModbusDeviceAndData(dev *sdk.Device) (*ModbusDeviceAndData, error) {
	//cfg, err := config.ModbusConfigFromDevice(dev)
	cfg, err := config.ModbusDeviceDataFromDevice(dev)
	if err != nil {
		return nil, err
	}
	return &ModbusDeviceAndData{
		Device: dev,
		Data:   cfg,
	}, nil
}


func LoadModbusDevice(plugin *sdk.Plugin, device *sdk.Device) (err error) {
  dev, err := NewModbusDeviceAndData(device)
  if err != nil {
    log.WithError(err).Error("failed to create new ModbusDeviceAndData"
    return
  }

  // TODO: Remainder.
}
*/

/*

// ModbusDeviceManager holds the information needed by the Modbus plugin to perform bulk
// read operations for configured devices.
//
// Having the ModbusDeviceManager as a higher-level abstraction above SDK devices allows
// us to aggregate the devices based on their common modbus configurations. This lets
// the plugin optimize reads by performing them in bulk. Instead of issuing a read call
// for each register for each device, a contiguous block of registers could be read at
// once, reducing the number of calls which need to be made.
//
// For convenience, this struct embeds the ModbusConfig struct, which generally
// contains all the pertinent connection configuration information specified by devices
// in their Data field.
type ModbusDeviceManager struct {
	config.ModbusConfig

	Blocks  []*ReadBlock
	Devices []*ModbusDevice

	// Internal flags denoting whether the devices have been sorted
	// and parsed into bulk read blocks.
	sorted bool
	parsed bool
}

// NewModbusDeviceManager creates a new instance of a ModbusDeviceManager from
// a given seed device. The particulars of the device are not used by the manager,
// but the device's Data field is used to fill in the ModbusConfig.
//
// Note: It is not within the purview of this function to check whether an existing
// ModbusDeviceManager exists for the given configuration. This responsibility is
// left to the caller.
func NewModbusDeviceManager(seed *ModbusDevice) (*ModbusDeviceManager, error) {
	if seed == nil {
		return nil, errors.New("unable to create new ModbusDeviceManager: seed device is nil")
	}

	if err := seed.Config.Validate(); err != nil {
		return nil, err
	}

	manager := &ModbusDeviceManager{
		ModbusConfig: *seed.Config,
		Devices:      []*ModbusDevice{seed},
		Blocks:       []*ReadBlock{},
	}
	return manager, nil
}

// MatchesDevice checks whether the ModbusDeviceManager "matches" a device. There
// is a match when the device's modbus configuration matches that of the manager.
func (d *ModbusDeviceManager) MatchesDevice(dev *ModbusDevice) bool {
	if dev == nil {
		return false
	}
	// TODO: determine whether all four of these data points are required to determine
	//   equality. Host and Port definitely are. Timeout and FailOnError seem less necessary
	//   for equality, but at the same time, if those values differ, it introduces some undefined
	//   behaviors for how things should behave if they need different failure trapping or timeouts.
	return d.Host == dev.Config.Host && d.Port == dev.Config.Port && d.Timeout == dev.Config.Timeout && d.FailOnError == dev.Config.FailOnError
}

// AddDevice adds a ModbusDevice to the manager to be tracked. It is the responsibility
// of the caller to ensure that the device being added matches the modbus config for the
// manager (MatchesDevice).
func (d *ModbusDeviceManager) AddDevice(dev *ModbusDevice) {
	if dev == nil {
		return
	}

	// Set internal flags indicating that the collection of devices needs to
	// be re-sorted and re-parsed.
	d.sorted = false
	d.parsed = false

	d.Devices = append(d.Devices, dev)
}

// Sort the devices managed by the ModbusDeviceManager. Sorting is done based on the
// device's modbus configuration, such as host, port, and register address.
func (d *ModbusDeviceManager) Sort() {
	sort.Sort(ByModbusConfig(d.Devices))

	// Set an internal flag indicating that the collection of devices has been
	// sorted.
	d.sorted = true
}

// ParseBlocks parses the ModbusDeviceManager's devices into blocks of registers for
// bulk reading. If the devices have already been parsed, this will do nothing. If
// another device has been added to the device manager since last parse, the devices
// will need to be re-sorted and re-pased.
//
// Parsing works by sorting the devices and calculating the number of registers between
// them. All devices whose registers fall within a maximum register count will be part
// of the same block.
func (d *ModbusDeviceManager) ParseBlocks() error {
	// If the blocks have already been parsed, there is nothing to do here.
	if d.parsed {
		return nil
	}

	// If the devices have not yet been sorted, we can not accurately parse them into
	// blocks for bulk read.
	if !d.sorted {
		return ErrDevicesNotSorted
	}

	log.Debug("parsing ModbusDeviceManager into bulk read blocks...")

	// If we get here, the devices have not yet been parsed into blocks and they have
	// been sorted, so they are ready to be parsed into blocks. A block does not need to
	// be contiguous (e.g. the registers do not need to immediately follow one another),
	// but they must all fall within a block the size of `MaximumRegisterCount`. Register
	// addresses/widths are used to calculate which devices fall into the block. If a
	// block exceeds the maximum register count, a new block will be started.
	var currentBlock *ReadBlock
	var sortIdx int32
	for _, dev := range d.Devices {
		// Increment the sort index set for the device.
		dev.Device.SortIndex = sortIdx
		sortIdx++

		if currentBlock == nil {
			currentBlock = NewReadBlock(dev)
			continue
		}

		// Address + Width calculates the largest register needed for this device. Subtracting
		// the StartRegister offsets this around 0 so we can check whether the total register block
		// size exceeds the maximum.
		newRegisterCount := (dev.Config.Address + dev.Config.Width) - currentBlock.StartRegister

		log.WithFields(log.Fields{
			"deviceAddress":    dev.Config.Address,
			"deviceWidth":      dev.Config.Width,
			"startRegister":    currentBlock.StartRegister,
			"registerCount":    currentBlock.RegisterCount,
			"newRegisterCount": newRegisterCount,
		}).Debug("calculating block envelope")

		// The new register is less than the max count, add the device to the current block.
		if newRegisterCount < MaximumRegisterCount {
			log.Debug("device within block bounds - adding to block")
			currentBlock.Add(dev)

		} else {
			log.Debug("device outside block bounds - creating new block")
			// The new register is over the max count. Store the current block and create a
			// new block starting with the current device.
			d.Blocks = append(d.Blocks, currentBlock)
			currentBlock = NewReadBlock(dev)
		}
	}
	// We are done generating Blocks; stash the last block that was added to into
	// the device manager Blocks field.
	d.Blocks = append(d.Blocks, currentBlock)

	log.WithField("blocks", len(d.Blocks)).Debug("successfully parsed read blocks")
	// We have now successfully parsed the devices into blocks suitable for bulk reads.
	d.parsed = true
	return nil
}

// NewClient creates a new modbus client using the manager's modbus configuration.
//
// A new client is created for each run of a BulkRead. The manager does not cache
// a client to prevent issues with long-lived connections and session resets.
func (d *ModbusDeviceManager) NewClient() (modbus.Client, error) {
	client, err := NewClient(&d.ModbusConfig)
	if err != nil {
		if d.FailOnError {
			return nil, err
		}
		log.WithField("error", err).Warning(
			"failed creating client when failOnError is disabled",
		)
	}
	return client, nil
}

// ReadBlock holds the information for a single block of registers for a bulk read.
type ReadBlock struct {
	Devices       []*ModbusDevice
	StartRegister uint16
	RegisterCount uint16
	Results       []byte
}

// NewReadBlock creates a new ReadBlock, using the provided device as a seed for the
// read block, which it inherits its start address and start register count from.
func NewReadBlock(seed *ModbusDevice) *ReadBlock {
	return &ReadBlock{
		Devices:       []*ModbusDevice{seed},
		StartRegister: seed.Config.Address,
		RegisterCount: seed.Config.Width,
		Results:       []byte{},
	}
}

// Add a modbus device to the ReadBlock. It is expected that the device being added
// has already been sorted. It is the responsibility of the caller to ensure this.
func (b *ReadBlock) Add(dev *ModbusDevice) {
	if dev == nil {
		return
	}
	b.Devices = append(b.Devices, dev)
	b.RegisterCount = (dev.Config.Address + dev.Config.Width) - b.StartRegister
}

// NewModbusClient creates a new modbus client from the given device configuration.
func NewModbusClient(device *sdk.Device) (modbus.Client, error) {
	var cfg config.ModbusConfig
	if err := mapstructure.Decode(device.Data, &cfg); err != nil {
		return nil, err
	}

	client, err := NewClient(&cfg)
	if err != nil {
		if cfg.FailOnError {
			return nil, err
		}
		log.WithField("error", err).Warning(
			"error creating client when failOnError is disabled",
		)
	}
	return client, nil
}

// UnpackRegisterReading creates a reading for the specified device using the device
// info and the device's modbus configuration as indices into the bulk read block results
// to get the reading value.
func UnpackRegisterReading(output *output.Output, block *ReadBlock, device *ModbusDevice) (*output.Reading, error) {
	startOffset := (2 * device.Config.Address) - (2 * block.StartRegister) // Results are in bytes, need 16-bit words.
	endOffset := startOffset + (2 * device.Config.Width)

	l := log.WithFields(log.Fields{
		"address":       device.Config.Address,
		"startRegister": block.StartRegister,
		"startOffset":   startOffset,
		"endOffset":     endOffset,
		"resultsLen":    len(block.Results),
	})
	l.Debug("unpacking register reading")

	if int(endOffset) > len(block.Results) {
		l.Error("failed bounds check when unpacking register reading")
		if device.Config.FailOnError {
			return nil, errors.New("bounds check failure during register read unpack")
		}
		l.Debug("not failing on error, returning nil device reading")
		return nil, nil // No reading
	}

	raw := block.Results[startOffset:endOffset]
	return UnpackReading(output, device, raw)
}

// UnpackCoilReading gets a coil reading (true / false) for the specified device from the
// bulk read block results bytes.
func UnpackCoilReading(output *output.Output, block *ReadBlock, device *ModbusDevice) (*output.Reading, error) {
	fmt.Printf("UnpackCoilReading start\n")
	fmt.Printf("output: %+v\n", output)
	fmt.Printf("block: %+v\n", block)
	fmt.Printf("device: %+v\n", device)
	bitIndex := device.Config.Address - block.StartRegister
	byteIndex := bitIndex / 8
	bitIndex = bitIndex % 8

	l := log.WithFields(log.Fields{
		"address":       device.Config.Address,
		"startRegister": block.StartRegister,
		"resultsLen":    len(block.Results),
		"bitIndex":      bitIndex,
		"byteIndex":     byteIndex,
	})
	l.Debug("unpacking coil reading")

	if int(byteIndex) >= len(block.Results) {
		l.Error("failed to get coil data: index out of bounds")
		if device.Config.FailOnError {
			return nil, errors.New("failed to get coil data: index out of bounds")
		}
		l.Debug("not failing on error, returning nil device reading")
		return nil, nil // No reading
	}

	coilByte := block.Results[byteIndex]
	mask := byte(0x01 << bitIndex)
	coilState := (coilByte & mask) != 0

	log.WithFields(log.Fields{
		"coilByte":  coilByte,
		"mask":      mask,
		"coilState": coilState,
	}).Debug("calculating coil state")

	return output.MakeReading(coilState), nil
}

// UnpackReading is a convenience wrapper for CastToType and MakeReading.
func UnpackReading(output *output.Output, device *ModbusDevice, reading []byte) (*output.Reading, error) {

	// Cast the reading bytes value to the specified type
	data, err := utils.CastToType(device.Config.Type, reading)
	if err != nil {
		l := log.WithFields(log.Fields{
			"data": reading,
			"type": device.Config.Type,
			"err":  err,
		})
		l.Error("failed to cast reading data to configured type")
		if device.Config.FailOnError {
			return nil, err
		}
		l.Debug("not failing on error, returning nil device reading")
		return nil, nil // No reading
	}

	return output.MakeReading(data), nil
}
*/
