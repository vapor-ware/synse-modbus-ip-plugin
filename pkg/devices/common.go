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
	"github.com/vapor-ware/synse-sdk/v2/sdk"
	"github.com/vapor-ware/synse-sdk/v2/sdk/output"
)

// MaximumRegisterCount is maximum numbr of registers in a modbus packetfor ReadHoldingRegisters.
const MaximumRegisterCount uint16 = 123

// GetModbusDeviceDataAndClient is common code to get the modbus configuration
// and client from the device configuration.
// handler is returned so that the caller can Close it.
func GetModbusDeviceDataAndClient(device *sdk.Device) (
	modbusDeviceData *config.ModbusDeviceData, client *modbus.Client, handler *modbus.TCPClientHandler, err error) {

	// Pull the modbus configuration out of the device Data fields.
	var deviceData config.ModbusDeviceData
	err = mapstructure.Decode(device.Data, &deviceData)
	if err != nil {
		return
	}

	// Create the modbus client from the configuration data.
	cli, handler, err := utils.NewClient(&deviceData)
	if err != nil {
		return
	}
	return &deviceData, &cli, handler, nil
}

// GetBulkReadClient gets the modbus client and device data for the
// connection information in k.
// handler is returned so that the caller can Close it.
func GetBulkReadClient(k ModbusBulkReadKey) (
	client modbus.Client, handler *modbus.TCPClientHandler, modbusDeviceData *config.ModbusDeviceData, err error) {
	log.Debugf("Creating modbus connection")
	modbusDeviceData = &config.ModbusDeviceData{
		Host:        k.Host,
		Port:        k.Port,
		Timeout:     k.Timeout,
		FailOnError: k.FailOnError,
		SlaveID:     k.SlaveID,
	}
	log.Debugf("modbusDeviceData: %#v", modbusDeviceData)
	client, handler, err = utils.NewClient(modbusDeviceData)
	if err != nil {
		log.Errorf("modbus NewClient failure: %v", err.Error())
		if modbusDeviceData.FailOnError {
			return
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
		// Make a reading with a nil Reading.Value.
		reading, _ = output.MakeReading(nil)
		if failOnErr {
			return reading, fmt.Errorf("failed to get coil data")
		}
		return reading, nil // Reading with nil reading.Value
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
func UnpackReading(output *output.Output, typeName string, rawReading []byte, failOnErr bool) (reading *output.Reading, err error) {

	// Cast the raw reading value to the specified output type
	data, err := utils.CastToType(typeName, rawReading)
	if err != nil {
		// Make a reading with a nil Reading.Value.
		reading, _ = output.MakeReading(nil)
		log.Errorf("Failed to cast typeName: %v, rawReading: %x", typeName, rawReading)
		if failOnErr {
			return reading, err
		}
		return reading, nil // No reading.
	}

	return output.MakeReading(data)
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
	// SlaveID is the modbus slave address which is not normally used in modbus over TCP.
	SlaveID int
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
// Returns sorted which is a slice of ModbusDevice in ascending register order.
// Returns deviceMap which is a map of register to sdk.Device.
func SortDevices(devices []*sdk.Device) (
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
			Register: deviceData.Address,
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
func MapBulkRead(devices []*sdk.Device, isCoil bool) (
	bulkReadMap map[ModbusBulkReadKey][]*ModbusBulkRead, keyOrder []ModbusBulkReadKey, err error) {

	log.Debugf("MapBulkRead start. devices: %+v", devices)
	for z := 0; z < len(devices); z++ {
		log.Debugf("MapBulkRead devices[%v]: %#v", z, devices[z])
	}

	// Sort the devices.
	sorted, sortedDevices, err := SortDevices(devices)
	if err != nil {
		log.Errorf("failed to sort devices")
		return nil, keyOrder, err
	}
	bulkReadMap = make(map[ModbusBulkReadKey][]*ModbusBulkRead)

	for z := 0; z < len(sorted); z++ {
		log.Debugf("MapBulkRead sorted[%v]: %#v", z, sorted[z])
	}

	for i := 0; i < len(sorted); i++ {
		// Create the key for this device from the device data.
		device := sortedDevices[sorted[i]]
		log.Debugf("--- next synse device: %v", device)
		var deviceData config.ModbusDeviceData
		err = mapstructure.Decode(device.Data, &deviceData)
		if err != nil {
			// Hard failure on configuration issue.
			log.Errorf(
				"MapBulkRead failed parsing device.Data, device at:[%v], device: %#v",
				i, device)
			return nil, keyOrder, err
		}

		key := ModbusBulkReadKey{
			Host:                 deviceData.Host,
			Port:                 deviceData.Port,
			Timeout:              deviceData.Timeout,
			FailOnError:          deviceData.FailOnError,
			SlaveID:              deviceData.SlaveID,
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

		deviceDataAddress := deviceData.Address
		deviceDataWidth := deviceData.Width

		log.Debugf("deviceDataAddress: 0x%04x", deviceDataAddress)
		log.Debugf("deviceDataWidth: %d", deviceDataWidth)

		// Insert.
		// If the key is not present, this is a simple insert to the map.
		if !keyPresent {
			log.Debugf("Key not present.")
			modbusBulkRead, err := NewModbusBulkRead(device, deviceDataAddress, deviceDataWidth, isCoil)
			if err != nil {
				return nil, keyOrder, err
			}
			log.Debugf("appending modbusBulkRead: %#v", modbusBulkRead)
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
			newRegisterCount := deviceDataAddress + deviceDataWidth - startRegister

			if newRegisterCount <= key.MaximumRegisterCount {
				log.Debugf("read fits in existing. newRegisterCount: %v", newRegisterCount)
				lastRead.RegisterCount = newRegisterCount
				lastRead.Devices = append(lastRead.Devices, device)
			} else {
				// Add a new read.
				log.Debugf("read does not fit in existing. newRegisterCount: %v", newRegisterCount)
				modbusBulkRead, err := NewModbusBulkRead(device, deviceDataAddress, deviceDataWidth, isCoil)
				if err != nil {
					return nil, keyOrder, err
				}
				log.Debugf("modbusBulkRead: %#v", modbusBulkRead)
				bulkReadMap[key] = append(bulkReadMap[key], modbusBulkRead)
			}
		}
	} // For each device.
	return bulkReadMap, keyOrder, nil
}

// DumpBulkReadMap dumps the map in key order to the log at Info.
func DumpBulkReadMap(bulkReadMap map[ModbusBulkReadKey][]*ModbusBulkRead, keyOrder []ModbusBulkReadKey) {

	for a := 0; a < len(keyOrder); a++ {
		k := keyOrder[a]
		v := bulkReadMap[k]

		// Dump the key.
		log.Infof("%#v", k)

		// Dump each read in v.
		for i := 0; i < len(v); i++ {
			log.Infof("    %d: StartRegister: %d, RegisterCount: %d", i, v[i].StartRegister, v[i].RegisterCount)
		}
	}
}

// MapBulkReadData maps the data read over modbus to the device read contexts.
func MapBulkReadData(bulkReadMap map[ModbusBulkReadKey][]*ModbusBulkRead, keyOrder []ModbusBulkReadKey) (
	readContexts []*sdk.ReadContext, err error) {
	// This map tells us if we have already created a read context for this
	// device and output. We can hit the same device and output more than once in
	// this loop when there are multiple modbus reads for a single device (more
	// than the maximum number ofregister addresses)
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

				// Get address and width.
				var deviceData config.ModbusDeviceData
				err := mapstructure.Decode(device.Data, &deviceData)
				if err != nil { // This is a configuration issue.
					log.Errorf(
						"MapBulkReadData failed parsing device at:[%v], device: %#v",
						i, device)
					return nil, err
				}

				deviceDataAddress := deviceData.Address
				deviceDataWidth := deviceData.Width

				log.Debugf("deviceDataAddress: 0x%04x", deviceDataAddress)
				log.Debugf("deviceDataWidth: %d", deviceDataWidth)
				log.Debugf("k.FailOnError: %v", k.FailOnError)

				readResults := read.ReadResults // Raw byte results from modbus call.

				var reading *output.Reading
				if read.IsCoil {
					reading, err = UnpackCoilReading(theOutput, read.ReadResults, read.StartRegister, deviceDataAddress, k.FailOnError)
					if err != nil {
						return nil, err
					}
				} else {
					// Get start and end data offsets. Bounds check.
					startDataOffset := (2 * deviceDataAddress) - (2 * read.StartRegister) // Results are in bytes. Need 16 bit words.
					endDataOffset := startDataOffset + (2 * deviceDataWidth)              // Two bytes per register.
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
						// Make a reading with a nil Reading.Value.
						reading, err = theOutput.MakeReading(nil)
						if err != nil {
							return nil, err
						}
						readings = append(readings, reading)
						// Append a read context here for the nil reading.
						readContext := sdk.NewReadContext(device, readings)
						readContexts = append(readContexts, readContext)
						continue // Next device.
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
				accountedFor[device] = theOutput

				// Only append a read context if we have readings. (Including nil readings)
				if len(readings) > 0 {
					readContext := sdk.NewReadContext(device, readings)
					readContexts = append(readContexts, readContext)
					log.Debugf("Appending readContext: %#v, device: %v", readContext, device)
				} else {
					log.Debugf("No readings to append. Not creating read context")
				}
			} // End for each device.
		} // End for each read.
	} // End for each key, value.
	return
}

// ModbusCallCounter is for testing. We increment it once after each network round
// trip with any modbus server.
var modbusCallCounter uint64
var callCounterMutex sync.Mutex

// GetModbusCallCounter gets the number of modbus calls to any modbus server.
func GetModbusCallCounter() (counter uint64) {
	callCounterMutex.Lock()
	counter = modbusCallCounter
	callCounterMutex.Unlock()
	return
}

// ResetModbusCallCounter resets the counter to zero for test purposes.
func ResetModbusCallCounter() {
	callCounterMutex.Lock()
	modbusCallCounter = 0
	callCounterMutex.Unlock()
}

// incrementModbusCallCounter is called internally whenever a modbus request is
// made to any modbus server.
func incrementModbusCallCounter() {
	callCounterMutex.Lock()
	if modbusCallCounter == math.MaxUint64 {
		modbusCallCounter = 0 // roll over
	} else {
		modbusCallCounter++
	}
	callCounterMutex.Unlock()
}

// bulkReadManager aggregates devices for bulk read.
type bulkReadManager struct {
	devices        []*sdk.Device // A slice of all devices.
	coilDevices    []*sdk.Device // A slice of all coil devices.
	holdingDevices []*sdk.Device // A slice of all holding register devices.
	inputDevices   []*sdk.Device // A slice of all input register devices.
	setupCompleted bool          // true once all setup is completed and we can perform bulk reads.

	coilBulkReadMap map[ModbusBulkReadKey][]*ModbusBulkRead // Mapped bulk reads for coils.
	coilKeyOrder    []ModbusBulkReadKey                     // Order of the keys to traverse the coilBulkReadMap.

	holdingBulkReadMap map[ModbusBulkReadKey][]*ModbusBulkRead // Mapped bulk reads for holding registers.
	holdingKeyOrder    []ModbusBulkReadKey                     // Order of the keys to traverse the holdingBulkReadMap.

	inputBulkReadMap map[ModbusBulkReadKey][]*ModbusBulkRead // Mapped bulk reads for input registers.
	inputKeyOrder    []ModbusBulkReadKey                     // Order of the keys to traverse the inputBulkReadMap.

	// true to make the read only coils bulk read a noop.
	// This will be false unless there are only read only coils and no read/write coils.
	shortOutReadOnlyCoil bool

	// true to make the read only holding registers bulk read a noop.
	// This will be false unless there are only read only holding registers and no read/write holding registers.
	shortOutReadOnlyHolding bool
}

// addModbusDevice adds a sdk.Device to a bulkReadManager.
func (brm *bulkReadManager) addModbusDevice(d *sdk.Device) (err error) {
	// Error check.
	if brm == nil {
		return fmt.Errorf("brm is nil")
	}
	if d == nil {
		return fmt.Errorf("d is nil")
	}

	// Append to the device slices.
	brm.devices = append(brm.devices, d)

	if d.Handler == "coil" || d.Handler == "read_only_coil" {
		if d.Handler == "coil" {
			brm.shortOutReadOnlyCoil = true
		}
		brm.coilDevices = append(brm.coilDevices, d)
		return
	}

	if d.Handler == "holding_register" || d.Handler == "read_only_holding_register" {
		if d.Handler == "holding_register" {
			brm.shortOutReadOnlyHolding = true
		}
		brm.holdingDevices = append(brm.holdingDevices, d)
		return
	}

	if d.Handler == "input_register" {
		brm.inputDevices = append(brm.inputDevices, d)
		return
	}

	return fmt.Errorf("Unknown device handler %s", d.Handler)
}

// bulkReadSetupMutex puts a critical section around bulkReadManager.setup() so
// that bulk read calls scheduled in parallel do not collide.
var bulkReadSetupMutex sync.Mutex

// setup sets up the manager for bulk read. If the manager is already setup, this is a noop.
func (brm *bulkReadManager) setup() (err error) {
	// Error check.
	if brm == nil {
		return fmt.Errorf("brm is nil")
	}

	bulkReadSetupMutex.Lock()
	if brm.setupCompleted {
		bulkReadSetupMutex.Unlock()
		return
	}

	log.Infof("Setting up bulk read")

	// Map out the bulk reads for coils.
	brm.coilBulkReadMap, brm.coilKeyOrder, err = MapBulkRead(brm.coilDevices, true)
	if err != nil {
		bulkReadSetupMutex.Unlock()
		return
	}
	log.Info("coilBulkReadMap:")
	DumpBulkReadMap(brm.coilBulkReadMap, brm.coilKeyOrder)

	// Map out the bulk reads for holding registers.
	brm.holdingBulkReadMap, brm.holdingKeyOrder, err = MapBulkRead(brm.holdingDevices, false)
	if err != nil {
		bulkReadSetupMutex.Unlock()
		return
	}
	log.Info("holdingBulkReadMap:")
	DumpBulkReadMap(brm.holdingBulkReadMap, brm.holdingKeyOrder)

	// Map out the bulk reads for input registers.
	brm.inputBulkReadMap, brm.inputKeyOrder, err = MapBulkRead(brm.inputDevices, false)
	if err != nil {
		bulkReadSetupMutex.Unlock()
		return
	}
	log.Info("inputBulkReadMap:")
	DumpBulkReadMap(brm.inputBulkReadMap, brm.inputKeyOrder)

	brm.setupCompleted = true
	log.Infof("Bulk read setup completed")

	log.Infof("shortOutReadOnlyCoil: %v\n", brm.shortOutReadOnlyCoil)
	log.Infof("shortOutReadOnlyHolding: %v\n", brm.shortOutReadOnlyHolding)
	bulkReadSetupMutex.Unlock()
	return
}

// GetBulkReadMap get the bulk read map and key order for the given mapId.
// Valid mapIds are coil, holding, input.
func (brm *bulkReadManager) GetBulkReadMap(mapID string) (
	bulkReadMap map[ModbusBulkReadKey][]*ModbusBulkRead, keyOrder []ModbusBulkReadKey, err error) {

	if brm == nil {
		err = fmt.Errorf("brm is nil")
		return
	}

	if mapID == "coil" {
		return brm.coilBulkReadMap, brm.coilKeyOrder, nil
	}

	if mapID == "holding" {
		return brm.holdingBulkReadMap, brm.holdingKeyOrder, nil
	}

	if mapID == "input" {
		return brm.inputBulkReadMap, brm.inputKeyOrder, nil
	}

	err = fmt.Errorf("Unknown mapId %s", mapID)
	return
}

// GetCoilsShortedOut returns true if BulkReadReadOnlyCoils should be a no-op.
func (brm *bulkReadManager) GetCoilsShortedOut() (shortedOut bool, err error) {
	if brm == nil {
		err = fmt.Errorf("brm is nil")
		return
	}
	return brm.shortOutReadOnlyCoil, nil
}

// GetHoldingShortedOut returns true if BulkReadReadOnlyHoldingRegisters should be a no-op.
func (brm *bulkReadManager) GetHoldingShortedOut() (shortedOut bool, err error) {
	if brm == nil {
		err = fmt.Errorf("brm is nil")
		return
	}
	return brm.shortOutReadOnlyHolding, nil
}

// brManager is a file level global that aggregates devices for bulk read.
var brManager bulkReadManager

// AddModbusDevice runs once during plugin initialization for each synse modbus device.
func AddModbusDevice(p *sdk.Plugin, d *sdk.Device) (err error) {
	return brManager.addModbusDevice(d)
}

// SetupBulkRead sets up the bulk read manager for bulk reads.
// If setup is already done, this is a noop.
func SetupBulkRead() {
	brManager.setup()
}

// GetBulkReadMap get the bulk read map and key order for the given mapId.
// Valid mapIds are coil, holding, input.
func GetBulkReadMap(mapID string) (
	bulkReadMap map[ModbusBulkReadKey][]*ModbusBulkRead, keyOrder []ModbusBulkReadKey, err error) {
	return brManager.GetBulkReadMap(mapID)
}

// GetCoilsShortedOut returns true if BulkReadReadOnlyCoils should be a no-op.
func GetCoilsShortedOut() (shortedOut bool, err error) {
	return brManager.GetCoilsShortedOut()
}

// GetHoldingShortedOut returns true if BulkReadReadOnlyCoils should be a no-op.
func GetHoldingShortedOut() (shortedOut bool, err error) {
	return brManager.GetHoldingShortedOut()
}

// PurgeBulkReadManager is a test only function to reset brManager.
func PurgeBulkReadManager() {
	log.Warn("Purging bulk read manager")
	brManager = bulkReadManager{}
}

// OnModbusDeviceLoad is a setup action which is called once per modbus device.
// This adds each synse modbus device to the bulkReadManager.
var OnModbusDeviceLoad = sdk.DeviceAction{
	Name:   "modbus-device-load",
	Filter: map[string][]string{"type": {"*"}}, // All devices
	Action: AddModbusDevice,
}
