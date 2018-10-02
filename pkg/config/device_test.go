package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// Get the timeout successfully.
func TestModbusDeviceData_GetTimeout(t *testing.T) {
	data := ModbusDeviceData{
		Timeout: "5s",
	}
	actual, err := data.GetTimeout()
	assert.NoError(t, err)
	assert.Equal(t, 5*time.Second, actual)
}

// Get an invalid timeout.
func TestModbusDeviceData_GetTimeout2(t *testing.T) {
	data := ModbusDeviceData{
		Timeout: "foobar",
	}
	_, err := data.GetTimeout()
	assert.Error(t, err)
}

// Validate successfully.
func TestModbusDeviceData_Validate(t *testing.T) {
	data := ModbusDeviceData{
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
	data := ModbusDeviceData{
		Port: 5000,
	}
	err := data.Validate()
	assert.Error(t, err)
}

// Invalid: No port
func TestModbusDeviceData_Validate3(t *testing.T) {
	data := ModbusDeviceData{
		Host: "localhost",
	}
	err := data.Validate()
	assert.Error(t, err)
}

// Valid: No timeout
func TestModbusDeviceData_Validate4(t *testing.T) {
	data := ModbusDeviceData{
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
