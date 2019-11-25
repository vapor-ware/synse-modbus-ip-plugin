package devices

import (
	"fmt"

	"github.com/goburrow/modbus"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
)

// NewClient gets a new Modbus client configured for TCP communication
// using the device's configuration.
func NewClient(data *config.ModbusDeviceData) (modbus.Client, error) {

	// Validate that the device config has all required fields.
	if err := data.Validate(); err != nil {
		return nil, err
	}

	// Parse the value for the client timeout.
	timeout, err := data.GetTimeout()
	if err != nil {
		return nil, err
	}

	// Create the TCP handler for the client
	handler := modbus.NewTCPClientHandler(fmt.Sprintf("%v:%v", data.Host, data.Port))
	handler.Timeout = timeout
	handler.SlaveId = uint8(data.SlaveID)

	client := modbus.NewClient(handler)
	return client, nil
}
