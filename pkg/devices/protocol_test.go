package devices

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
)

func TestNewClient(t *testing.T) {
	cfg := config.ModbusDeviceData{
		Host: "localhost",
		Port: 7777,
	}

	client, err := NewClient(&cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)

}

func TestNewClient_ValidationError(t *testing.T) {
	cfg := config.ModbusDeviceData{
		// missing field: host
		Port: 7777,
	}

	client, err := NewClient(&cfg)
	assert.Error(t, err)
	assert.Nil(t, client)
}

func TestNewClient_InvalidTimeoutError(t *testing.T) {
	cfg := config.ModbusDeviceData{
		Host:    "localhost",
		Port:    7777,
		Timeout: "not-a-duration-string",
	}

	client, err := NewClient(&cfg)
	assert.Error(t, err)
	assert.Nil(t, client)
}
