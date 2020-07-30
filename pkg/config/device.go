package config

import (
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/vapor-ware/synse-sdk/sdk"
)

// ModbusDeviceData models the scheme for the supported config values
// of device's Data field for the Modbus TCP/IP plugin.
//// ModbusDeviceData is a device's modbus configuration. This is generally set
//// in the SDK Device's Data field.
//type ModbusDeviceData struct {
type ModbusDeviceData struct {
	// Host is the hostname/ip of the device to connect to.
	Host string `yaml:"host,omitempty"`

	// Port is the port number for the device.
	Port int `yaml:"port,omitempty"`

	// SlaveID is the modbus slave id.
	SlaveID int `yaml:"slaveId,omitempty"`

	// Timeout is the duration to wait for a modbus request to resolve.
	// FIXME: could the type here just be a duration? That may parse it automatically?
	//   need to check the capabilities of the mapstructure package
	Timeout string `yaml:"timeout,omitempty"`

	// FailOnError will cause a read to fail (e.g. return an error) if
	// any of the device fail to read in bulk. When failOnError is not
	// set, the error will typically only be logged. This is false by default.
	FailOnError bool `yaml:"failOnError,omitempty"`

	// We many or may not need these (unclear).

	// Address is the register address which holds the reading value.
	Address uint16

	// Width is the number of registers to read, starting from the `Address`.
	Width uint16

	// Type is the type of the data held in the registers. The supported
	// types are as follows:
	//
	//  "b", "bool", "boolean": boolean
	//  "u16", "uint16":  unsigned 16-bit integer
	//  "u32", "uint32":  unsigned 32-bit integer
	//  "u64", "uint64":  unsigned 64-bit integer
	//  "s16", "int16":   signed 16-bit integer
	//  "s32", "int32":   signed 32-bit integer
	//  "s64", "int64":   signed 64-bit integer
	//  "f32", "float32": 32-bit floating point number
	//  "f64", "float64": 64-bit floating point number
	Type string
}

// ModbusDeviceDataFromDevice creates a new instance of a ModbusDeviceData and loads
// it with values from the provided SDK Device's Data field.
func ModbusDeviceDataFromDevice(device *sdk.Device) (*ModbusDeviceData, error) {
	var cfg ModbusDeviceData
	if err := mapstructure.Decode(device.Data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetTimeout gets the timeout configuration as a duration.
func (data *ModbusDeviceData) GetTimeout() (time.Duration, error) {
	return time.ParseDuration(data.Timeout)
}

// Validate makes sure that the ModbusDeviceData instance has all of its
// required fields set.
func (data *ModbusDeviceData) Validate() error {
	if data.Host == "" {
		return fmt.Errorf("'host' not found in device config, %v", data)
	}
	if data.Port == 0 {
		return fmt.Errorf("'port' not found in device config %v", data)
	}
	if data.Timeout == "" {
		// If there is no timeout set, default to 5s
		data.Timeout = "5s"
	}
	return nil
}
