package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-sdk/sdk"
)

func TestModbusConfigFromDevice(t *testing.T) {
	d := &sdk.Device{
		Data: map[string]interface{}{
			"host":        "localhost",
			"port":        5050,
			"slaveId":     10,
			"timeout":     "10s",
			"failOnError": true,
			"address":     5,
			"width":       2,
			"type":        "u32",
		},
	}

	cfg, err := ModbusConfigFromDevice(d)
	assert.NoError(t, err)
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 5050, cfg.Port)
	assert.Equal(t, 10, cfg.SlaveID)
	assert.Equal(t, "10s", cfg.Timeout)
	assert.Equal(t, true, cfg.FailOnError)
	assert.Equal(t, uint16(5), cfg.Address)
	assert.Equal(t, uint16(2), cfg.Width)
	assert.Equal(t, "u32", cfg.Type)
}

func TestModbusConfigFromDevice_Error(t *testing.T) {
	d := &sdk.Device{
		Data: map[string]interface{}{
			"host": "localhost",
			"port": "5050", // should be int
		},
	}

	cfg, err := ModbusConfigFromDevice(d)
	assert.Error(t, err)
	assert.Nil(t, cfg)

}

// Get the timeout successfully.
func TestModbusDeviceData_GetTimeout(t *testing.T) {
	data := ModbusConfig{
		Timeout: "5s",
	}
	actual, err := data.GetTimeout()
	assert.NoError(t, err)
	assert.Equal(t, 5*time.Second, actual)
}

// Get an invalid timeout.
func TestModbusDeviceData_GetTimeout2(t *testing.T) {
	data := ModbusConfig{
		Timeout: "foobar",
	}
	_, err := data.GetTimeout()
	assert.Error(t, err)
}

// Validate successfully.
func TestModbusDeviceData_Validate(t *testing.T) {
	data := ModbusConfig{
		Host:    "localhost",
		Port:    5000,
		Timeout: "10s",
	}
	err := data.Validate()
	assert.NoError(t, err)

	assert.Equal(t, "localhost", data.Host)
	assert.Equal(t, 5000, data.Port)
	assert.Equal(t, 0, data.SlaveID)
	assert.Equal(t, "10s", data.Timeout)
	assert.Equal(t, false, data.FailOnError)
}

// Invalid: No host
func TestModbusDeviceData_Validate2(t *testing.T) {
	data := ModbusConfig{
		Port: 5000,
	}
	err := data.Validate()
	assert.Error(t, err)
}

// Invalid: No port
func TestModbusDeviceData_Validate3(t *testing.T) {
	data := ModbusConfig{
		Host: "localhost",
	}
	err := data.Validate()
	assert.Error(t, err)
}

// Valid: No timeout
func TestModbusDeviceData_Validate4(t *testing.T) {
	data := ModbusConfig{
		Host: "localhost",
		Port: 5000,
	}
	err := data.Validate()
	assert.NoError(t, err)

	assert.Equal(t, "localhost", data.Host)
	assert.Equal(t, 5000, data.Port)
	assert.Equal(t, 0, data.SlaveID)
	assert.Equal(t, "5s", data.Timeout)
	assert.Equal(t, false, data.FailOnError)
}
