package utils

import (
	"fmt"
	"time"

	"github.com/goburrow/modbus"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
)

// NewClient gets a new Modbus client configured for TCP communication using the
// device's configuration. handler is returned here so that the caller can Close
// it when done.
func NewClient(data *config.ModbusDeviceData) (
	client modbus.Client, handler *modbus.TCPClientHandler, err error) {

	// Validate that the device config has all required fields.
	err = data.Validate()
	if err != nil {
		return
	}

	// Parse the value for the client timeout.
	var timeout time.Duration
	timeout, err = data.GetTimeout()
	if err != nil {
		return
	}

	// Create the TCP handler for the client
	handler = modbus.NewTCPClientHandler(fmt.Sprintf("%v:%v", data.Host, data.Port))
	handler.Timeout = timeout
	handler.SlaveId = uint8(data.SlaveID)

	client = modbus.NewClient(handler)
	return
}
