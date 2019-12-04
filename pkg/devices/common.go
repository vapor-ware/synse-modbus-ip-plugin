package devices

// This file contains common modbus device code.
import (
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
)

var ErrDevicesNotSorted = errors.New("devices not sorted; unable to parse bulk read blocks")

// MaximumRegisterCount is the max number of registers to read in one modbus
// call. We may need to tune this for some devices (not clear). Technical max
// is 123 for ReadHoldingRegisters over IP.
const MaximumRegisterCount uint16 = 123

// ModbusDevice wraps an SDK Device and associates it with a ModbusConfig
// configuration parsed from the SDK Device's Data field.
type ModbusDevice struct {
	Device *sdk.Device
	Config *config.ModbusConfig
}

// NewModbusDevice creates a new instance of the ModbusDevice wrapper for the
// given SDK Device.
func NewModbusDevice(dev *sdk.Device) (*ModbusDevice, error) {
	cfg, err := config.ModbusConfigFromDevice(dev)
	if err != nil {
		return nil, err
	}
	return &ModbusDevice{
		Device: dev,
		Config: cfg,
	}, nil
}

// ModbusDeviceManager holds the information needed by the Modbus plugin to perform bulk
// read operations for configured devices.
//
// Having the ModbusDeviceManager as a higher-level abstraction above SDK devices allows
// us to aggregate the devices based on their common modbus configurations. This enables
// the plugin to try and optimize reads by performing them in bulk. Instead of issuing
// a read call for each register for each device, a contiguous block of registers could
// be read at once, reducing the number of calls which need to be made.
//
// For convenience, this struct embeds the ModbusConfig struct, which generally
// contains all the pertinent connection configuration information specified by devices
// in their Data field.
type ModbusDeviceManager struct {
	config.ModbusConfig

	Client  modbus.Client
	Blocks  []*ReadBlock
	Devices []*ModbusDevice

	// Internal flags denoting whether the devices have been sorted
	// and parsed into bulk read blocks.
	sorted bool
	parsed bool
}

// MatchesDevice
func (d *ModbusDeviceManager) MatchesDevice(dev *ModbusDevice) bool {
	// TODO: determine whether all four of these data points are required to determine
	//   equality. Host and Port definitely are. Timeout and FailOnError seem less necessary
	//   for equality, but at the same time, if those values differ, it introduces some undefined
	//   behaviors for how things should behave if they need different failure trapping or timeouts.
	return d.Host == dev.Config.Host && d.Port == dev.Config.Port && d.Timeout == dev.Config.Timeout && d.FailOnError == dev.Config.FailOnError
}

// AddDevice
func (d *ModbusDeviceManager) AddDevice(dev *ModbusDevice) {
	// Set internal flags indicating that the collection of devices needs to
	// be re-sorted and re-parsed.
	d.sorted = false
	d.parsed = false

	d.Devices = append(d.Devices, dev)
}

// Sort
func (d *ModbusDeviceManager) Sort() {
	sort.Sort(ByModbusConfig(d.Devices))

	// Set an internal flag indicating that the collection of devices has been
	// sorted.
	d.sorted = true
}

// ParseBlocks
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

		// The new register is less than the max count, add the device to the current block.
		if newRegisterCount < MaximumRegisterCount {
			currentBlock.Add(dev)

		} else {
			// The new register is over the max count. Store the current block and create a
			// new block starting with the current device.
			d.Blocks = append(d.Blocks, currentBlock)
			currentBlock = NewReadBlock(dev)
		}
	}

	// We have now successfully parsed the devices into blocks suitable for bulk reads.
	d.parsed = true
	return nil
}

// ReadBlock
type ReadBlock struct {
	Devices       []*ModbusDevice
	StartRegister uint16
	RegisterCount uint16
	Results       []byte
}

// NewReadBlock
func NewReadBlock(seed *ModbusDevice) *ReadBlock {
	return &ReadBlock{
		Devices:       []*ModbusDevice{seed},
		StartRegister: seed.Config.Address,
		RegisterCount: seed.Config.Width,
		Results:       []byte{},
	}
}

// Add
func (b *ReadBlock) Add(dev *ModbusDevice) {
	b.Devices = append(b.Devices, dev)
	b.RegisterCount = (dev.Config.Address + dev.Config.Width) - b.StartRegister
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

	manager := &ModbusDeviceManager{
		ModbusConfig: *seed.Config,
		Devices:      []*ModbusDevice{seed},
		Blocks:       []*ReadBlock{},
	}
	c, err := newModbusClientFromManager(manager)
	if err != nil {
		return nil, err
	}
	manager.Client = c

	return manager, nil
}

func newModbusClientFromManager(manager *ModbusDeviceManager) (modbus.Client, error) {
	client, err := NewClient(&manager.ModbusConfig)
	if err != nil {
		if manager.FailOnError {
			return nil, err
		}
		log.WithField("error", err).Warning(
			"failed creating client when failOnError is disabled",
		)
	}
	return client, nil
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

//// GetModbusClientAndConfig is common code to get the modbus configuration and client from the device configuration.
//func GetModbusClientAndConfig(device *sdk.Device) (modbusConfig *config.ModbusConfig, client modbus.Client, err error) {
//
//	// Pull the modbus configuration out of the device Data fields.
//	var deviceData config.ModbusConfig
//	err = mapstructure.Decode(device.Data, &deviceData)
//	if err != nil {
//		return nil, nil, err
//	}
//
//	// Create the modbus client from the configuration data.
//	cli, err := utils.NewClient(&deviceData)
//	if err != nil {
//		return nil, nil, err
//	}
//	return &deviceData, cli, nil
//}

//// GetBulkReadClient gets the modbus client and device data for the
//// connection information in k.
//func GetBulkReadClient(k ModbusBulkReadKey) (client modbus.Client, modbusDeviceData *config.ModbusConfig, err error) {
//	log.Debugf("Creating modbus connection")
//	modbusDeviceData = &config.ModbusConfig{
//		Host:        k.Host,
//		Port:        k.Port,
//		Timeout:     k.Timeout,
//		FailOnError: k.FailOnError,
//		// Omitting SlaveID for now. Not currently used.
//	}
//	log.Debugf("modbusDeviceData: %#v", modbusDeviceData)
//	client, err = NewClient(modbusDeviceData)
//	if err != nil {
//		log.Debugf("modbus NewClient failure: %v", err.Error())
//		if modbusDeviceData.FailOnError {
//			return nil, nil, err
//		}
//	}
//	return
//}

func UnpackRegisterReading(output *output.Output, rawReading []byte, startAddress, deviceAddress, deviceWidth uint16, typeName string, failOnErr bool) (*output.Reading, error) {
	startOffset := (2 * deviceAddress) - (2 * startAddress) // Results are in bytes, need 16-bit words.
	endOffset := startOffset + (2 * deviceWidth)

	if int(endOffset) > len(rawReading) {
		if failOnErr {
			return nil, fmt.Errorf("bounds check failure")
		}
		// Nil reading is returned if we are not to fail on error.
		return nil, nil
	}

	raw := rawReading[startOffset:endOffset]
	return UnpackReading(output, typeName, raw, failOnErr)
}

// UnpackCoilReading gets a coil (true / false) from a ReadCoils result buffer.
func UnpackCoilReading(output *output.Output, rawReading []byte, startAddress, coilAddress uint16, failOnErr bool) (
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

	reading = output.MakeReading(data)
	return
}

//// ModbusBulkReadKey corresponds to a Modbus Device / Connection.
//// We will need one or more bulk reads per key entry.
//type ModbusBulkReadKey struct {
//	// Modbus device host name.
//	Host string
//	// Modbus device port.
//	Port int
//	// Timeout for modbus read.
//	Timeout string
//	// Fail on error. (Do we abort on one failed read?)
//	FailOnError bool
//	// Maximum number of registers to read on a single modbus call to the device.
//	MaximumRegisterCount uint16
//}

//// NewModbusBulkReadKey creates a modbus bulk read key.
//func NewModbusBulkReadKey(host string, port int, timeout string, failOnError bool) (key *ModbusBulkReadKey, err error) {
//	if host == "" {
//		return nil, fmt.Errorf("empty host")
//	}
//	if port <= 0 {
//		return xnil, fmt.Errorf("invalid port %v", port)
//	}
//	key = &ModbusBulkReadKey{
//		Host:                 host,
//		Port:                 port,
//		Timeout:              timeout,
//		FailOnError:          failOnError,
//		MaximumRegisterCount: MaximumRegisterCount,
//	}
//	return
//}

//// ModbusBulkRead contains data for each individual bulk read call to the modbus device.
//type ModbusBulkRead struct {
//	// Synse devices associated with this read.
//	Devices []*sdk.Device
//	// Raw Modbus read results
//	ReadResults []byte
//	// First register to read.
//	StartRegister uint16
//	// Number of registers to read.
//	RegisterCount uint16
//	// true for coils. The unmarshalling is different.
//	IsCoil bool
//}

//// NewModbusBulkRead contains data for each bulk read.
//func NewModbusBulkRead(device *sdk.Device, startRegister, registerCount uint16, isCoil bool) (*ModbusBulkRead, error) {
//	if device == nil {
//		return nil, fmt.Errorf("no device pointer given")
//	}
//	return &ModbusBulkRead{
//		Devices:       []*sdk.Device{device},
//		StartRegister: startRegister,
//		RegisterCount: registerCount,
//		IsCoil:        isCoil,
//	}, nil
//}

//// ModbusDeviceOrig is an intermediate struct for sorting ModbusBulkReadKey.
//// TODO: private?
//type ModbusDeviceOrig struct {
//	Host     string
//	Port     int
//	Register uint16
//}

//// SortDevices sorts the device list.
//// Used for bulk reads.
//// Returns sorted which is a slice of ModbusDeviceOrig in ascending order.
//// Returns deviceMap which is a map of register to sdk.Device.
//func SortDevices(devices []*sdk.Device, setSortOrdinal bool) (
//	sorted []ModbusDeviceOrig, deviceMap map[ModbusDeviceOrig]*sdk.Device, err error) {
//
//	if devices == nil {
//		return nil, nil, nil // Nothing to sort. Could arguably fail here.
//	}
//	deviceMap = make(map[ModbusDeviceOrig]*sdk.Device)
//
//	// For each device.
//	for i := 0; i < len(devices); i++ {
//		device := devices[i]
//
//		// Deserialize the modbus configuration.
//		var deviceData config.ModbusConfig
//		err = mapstructure.Decode(device.Data, &deviceData)
//		if err != nil {
//			return nil, nil, err
//		}
//
//		key := ModbusDeviceOrig{
//			Host:     deviceData.Host,
//			Port:     deviceData.Port,
//			Register: uint16(deviceData.Address), // TODO: Can we have uint16 in the config struct.
//		}
//
//		// Add to locals.
//		sorted = append(sorted, key)
//		deviceMap[key] = device
//	} // end for each device
//
//	// Sort / trace.
//	sort.SliceStable(sorted, func(i, j int) bool {
//		if sorted[i].Host < sorted[j].Host {
//			return true
//		} else if sorted[i].Host > sorted[j].Host {
//			return false
//		}
//		if sorted[i].Port < sorted[j].Port {
//			return true
//		} else if sorted[i].Port > sorted[j].Port {
//			return false
//		}
//		if sorted[i].Register < sorted[j].Register {
//			return true
//		} else if sorted[i].Register > sorted[j].Register {
//			return false
//		}
//		log.Errorf("Duplicate modbus device configured. Host: %v, Port: %v, Register: %v",
//			sorted[i].Host, sorted[i].Port, sorted[i].Register)
//		return true
//	})
//
//	// Add SortOrdinal to all devices.
//	if setSortOrdinal {
//		for k := 0; k < len(sorted); k++ {
//			deviceMap[sorted[k]].SortIndex = NextSortOrdinal
//			NextSortOrdinal++
//		}
//	}
//	return
//}

//// MapBulkRead maps the physical modbus device / connection information for all
//// modbus devices to a map of each modbus bulk read call required to get all
//// register data configured for the device.
//func MapBulkRead(devices []*sdk.Device, setSortOrdinal bool, isCoil bool) (
//	bulkReadMap map[ModbusBulkReadKey][]*ModbusBulkRead, keyOrder []ModbusBulkReadKey, err error) {
//
//	//log.Debugf("MapBulkRead start. devices: %+v", devices)
//	//for z := 0; z < len(devices); z++ {
//	//	log.Debugf("MapBulkRead devices[%v]: %#v", z, devices[z])
//	//}
//
//	// Sort the devices based on their host/port/register. This should align devices
//	// based on sequential contiguous register blocks, if there are any.
//	//sort.Sort(ByModbusConfig(devices))
//
//	// If configured to set the sort ordinal on the devices, do so now.
//	if setSortOrdinal {
//		for _, device := range devices {
//			device.SortIndex = NextSortOrdinal
//			NextSortOrdinal++
//		}
//	}
//
//	// Sort the devices.
//	sorted, sortedDevices, err := SortDevices(devices, setSortOrdinal)
//	if err != nil {
//		log.Errorf("failed to sort devices")
//		return nil, keyOrder, err
//	}
//	bulkReadMap = make(map[ModbusBulkReadKey][]*ModbusBulkRead)
//
//	//for z := 0; z < len(sorted); z++ {
//	//	log.Debugf("MapBulkRead sorted[%v]: %#v", z, sorted[z])
//	//}
//
//	for i := 0; i < len(sorted); i++ {
//		// Create the key for this device from the device data.
//		device := sortedDevices[sorted[i]]
//		log.Debugf("--- next synse device: %v", device)
//		var deviceData config.ModbusConfig
//		err = mapstructure.Decode(device.Data, &deviceData)
//		if err != nil {
//			return nil, keyOrder, err
//		}
//
//		key := ModbusBulkReadKey{
//			Host:                 deviceData.Host,
//			Port:                 deviceData.Port,
//			Timeout:              deviceData.Timeout,
//			FailOnError:          deviceData.FailOnError,
//			MaximumRegisterCount: MaximumRegisterCount,
//		}
//		log.Debugf("Created key: %#v", key)
//
//		// Find out if the key is in the map.
//		keyValues, keyPresent := bulkReadMap[key]
//		if keyPresent {
//			log.Debugf("key is already in the map")
//		} else {
//			log.Debugf("key is not in the map")
//		}
//
//		log.Debugf("len(keyValues): %v", len(keyValues))
//
//		outputDataAddress := uint16(deviceData.Address) // TODO: Can we have uint16 in the config struct?
//		outputDataWidth := uint16(deviceData.Width)     // TODO: As above.
//
//		log.Debugf("outputDataAddress: 0x%04x", outputDataAddress)
//		log.Debugf("outputDataWidth: %d", outputDataWidth)
//
//		// Insert.
//		// If the key is not present, this is a simple insert to the map.
//		if !keyPresent {
//			log.Debugf("Key not present.")
//			modbusBulkRead, err := NewModbusBulkRead(device, outputDataAddress, outputDataWidth, isCoil)
//			if err != nil {
//				return nil, keyOrder, err
//			}
//			log.Debugf("modbusBulkRead: %#v", modbusBulkRead)
//			bulkReadMap[key] = append(bulkReadMap[key], modbusBulkRead)
//			keyOrder = append(keyOrder, key)
//		} else {
//			log.Debugf("Key present")
//			// See if this fits on the end of the slice.
//			//  If so, update the ModbusBulkRead RegisterCount.
//			//  If not, create a new ModbusBulkRead.
//			reads := bulkReadMap[key]
//			lastRead := reads[len(reads)-1]
//			startRegister := lastRead.StartRegister
//			log.Debugf("startRegister: 0x%0x", startRegister)
//			newRegisterCount := outputDataAddress + outputDataWidth - startRegister
//
//			if newRegisterCount < key.MaximumRegisterCount {
//				log.Debugf("read fits in existing. newRegisterCount: %v", newRegisterCount)
//				lastRead.RegisterCount = newRegisterCount
//				lastRead.Devices = append(lastRead.Devices, device)
//			} else {
//				// Add a new read.
//				log.Debugf("read does not fit in existing. newRegisterCount: %v", newRegisterCount)
//				modbusBulkRead, err := NewModbusBulkRead(device, outputDataAddress, outputDataWidth, isCoil)
//				if err != nil {
//					return nil, keyOrder, err
//				}
//				log.Debugf("modbusBulkRead: %#v", modbusBulkRead)
//				bulkReadMap[key] = append(bulkReadMap[key], modbusBulkRead)
//				keyOrder = append(keyOrder, key)
//			}
//		}
//	} // For each device.
//	return bulkReadMap, keyOrder, nil
//}

//// MapBulkReadData maps the data read over modbus to the device read contexts.
//func MapBulkReadData(bulkReadMap map[ModbusBulkReadKey][]*ModbusBulkRead, keyOrder []ModbusBulkReadKey) (
//	readContexts []*sdk.ReadContext, err error) {
//	// This map tells us if we have already created a read context for this
//	// device and output. We can hit the same device and output more than once in
//	// this loop when there are multiple modbus reads for a single device (more
//	// than 123 register addresses)
//	accountedFor := make(map[*sdk.Device][]*output.Output)
//
//	for a := 0; a < len(keyOrder); a++ {
//		k := keyOrder[a]
//		v := bulkReadMap[k]
//		for h := 0; h < len(v); h++ { // for each read
//			read := v[h]
//			devices := read.Devices
//
//			for i := 0; i < len(devices); i++ {
//				device := devices[i]
//
//				// For each device output.
//				outputs := []*output.Output{output.Get(device.Output)}
//				readings := []*output.Reading{}
//				for j := 0; j < len(outputs); j++ {
//					out := outputs[j]
//
//					// Have we accounted for this device and output yet?
//					// This can happen when multiple reads are required for a single ModbusBulkReadKey.
//					_, keyPresent := accountedFor[device]
//					inMap := false
//					if keyPresent {
//						// Device is there. Is the output there?
//						for b := 0; b < len(accountedFor[device]); b++ {
//							if accountedFor[device][b] == out {
//								inMap = true
//								break // for
//							}
//						}
//						if inMap {
//							log.Debugf("device[output] already accounted for: device %p, output %p", device, out)
//							continue // next output
//						}
//					}
//
//					var outputData config.ModbusConfig
//					// Get the output data. Need address and width.
//					err := mapstructure.Decode(device.Data, &outputData)
//					if err != nil { // This is not a configuration issue. Device may not have responded.
//						log.Errorf(
//							"MapBulkReadData failed parsing output.Data device at:[%v], device: %#v, output: %#v",
//							i, device, out)
//						if k.FailOnError {
//							return nil, err
//						}
//					}
//					outputDataAddress := uint16(outputData.Address) // TODO: Can we have uint16 in the config struct?
//					outputDataWidth := uint16(outputData.Width)     // TODO: As above.
//
//					log.Debugf("outputDataAddress: 0x%04x", outputDataAddress)
//					log.Debugf("outputDataWidth: %d", outputDataWidth)
//					log.Debugf("k.FailOnError: %v", k.FailOnError)
//
//					readResults := read.ReadResults // Raw byte results from modbus call.
//
//					var reading *output.Reading
//					if read.IsCoil {
//						reading, err = UnpackCoilReading(out, read.ReadResults, read.StartRegister, outputDataAddress, k.FailOnError)
//						if err != nil {
//							return nil, err
//						}
//					} else {
//						// Get start and end data offsets. Bounds check.
//						startDataOffset := (2 * outputDataAddress) - (2 * read.StartRegister) // Results are in bytes. Need 16 bit words.
//						endDataOffset := startDataOffset + (2 * outputDataWidth)              // Two bytes per register.
//						readResultsLength := len(readResults)
//
//						log.Debugf("startDataOffset: %d", startDataOffset)
//						log.Debugf("endDataOffset: %d", endDataOffset)
//						log.Debugf("readResultsLength: %d", readResultsLength)
//
//						if int(endDataOffset) > len(readResults) {
//							if k.FailOnError {
//								return nil, fmt.Errorf("bounds check failure. startDataOffset: %v, endDataOffset: %v, readResultsLength: %v",
//									startDataOffset, endDataOffset, readResultsLength)
//							}
//							// nil reading.
//							log.Errorf("No data. Attempt to read beyond bounds. startDataOffset: %v, endDataOffset: %v, readResultsLength: %v",
//								startDataOffset, endDataOffset, readResultsLength)
//							readings = append(readings, nil)
//							continue // Next output.
//						} // end bounds check.
//
//						rawReading := readResults[startDataOffset:endDataOffset]
//						log.Debugf("rawReading: len: %v, %x", len(rawReading), rawReading)
//
//						reading, err = UnpackReading(out, outputData.Type, rawReading, k.FailOnError)
//						if err != nil {
//							return nil, err
//						}
//					}
//					log.Debugf("Appending reading: %#v, device: %v, output: %#v", reading, device, out)
//					readings = append(readings, reading)
//
//					// Add to accounted for.
//					accountedFor[device] = append(accountedFor[device], out)
//
//				} // End for each output.
//
//				// Only append a read context if we have readings. (Including nil readings)
//				if len(readings) > 0 {
//					readContext := sdk.NewReadContext(device, readings)
//					readContexts = append(readContexts, readContext)
//					log.Debugf("Appending readContext: %#v, device: %v", readContext, device)
//				} else {
//					log.Debugf("No readings to append. Not creating read context")
//				}
//			} // End for each device.
//		} // End for each read.
//	} // End for each key, value.
//	return
//}
