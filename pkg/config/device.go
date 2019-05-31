package config

import (
	"fmt"
	"time"
)

// ModbusDeviceData models the scheme for the supported config values
// of device's Data field for the Modbus TCP/IP plugin.
type ModbusDeviceData struct {
	// Host is the hostname/ip of the device to connect to.
	Host string `yaml:"host,omitempty"`

	// Port is the port number for the device.
	Port int `yaml:"port,omitempty"`

	// SlaveID is the modbus slave id.
	SlaveID int `yaml:"slaveId,omitempty"`

	// Timeout is the duration to wait for a modbus request to resolve.
	Timeout string `yaml:"timeout,omitempty"`

	// FailOnError will cause a read to fail (e.g. return an error) if
	// any of the device's outputs fails to read. When failOnError is not
	// set, the error will only be logged. This is false by default.
	FailOnError bool `yaml:"failOnError,omitempty"`

	// Address is the register address which holds the output reading.
	Address int

	// Width is the number of registers to read, starting from the `Address`.
	Width int

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
