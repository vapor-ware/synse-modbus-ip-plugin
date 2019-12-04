package devices

// This file contains common modbus device code.
import (
	"errors"
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
	// We are done generating Blocks; stash the last block that was added to into
	// the device manager Blocks field.
	d.Blocks = append(d.Blocks, currentBlock)

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

// UnpackCoilReading gets a coil (true / false) from a ReadCoils result buffer.
func UnpackCoilReading(output *output.Output, block *ReadBlock, device *ModbusDevice) (*output.Reading, error) {
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

// UnpackReading is a wrapper for CastToType and MakeReading.
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
