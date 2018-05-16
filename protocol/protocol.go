package protocol

import (
	"fmt"
	"strconv"
	"time"

	"github.com/goburrow/modbus"
)

// NewClient gets a new Modbus client configured for TCP communication
// using the device's configuration.
func NewClient(config map[string]string) (modbus.Client, error) {

	// TODO (etd) -- there is a better way of doing this checking, but
	// because we want to get it work and because it will be changing
	// relatively soon due to SDK changes, this should be fine for now.
	var err error

	host, ok := config["host"]
	if !ok {
		return nil, fmt.Errorf("'host' not found in device config: %+v", config)
	}

	port, ok := config["port"]
	if !ok {
		return nil, fmt.Errorf("'port' not found in device config: %+v", config)
	}

	var slaveID uint8
	sid, ok := config["slave_id"]
	if !ok {
		return nil, fmt.Errorf("'slave_id' not found in device config: %+v", config)
	}
	var i int64
	i, err = strconv.ParseInt(sid, 10, 8)
	if err != nil {
		return nil, err
	}
	slaveID = uint8(i)

	var timeout time.Duration
	duration, ok := config["timeout"]
	if !ok {
		timeout = 5 * time.Second
	} else {
		timeout, err = time.ParseDuration(duration)
		if err != nil {
			return nil, err
		}
	}

	address := fmt.Sprintf("%s:%s", host, port)

	handler := modbus.NewTCPClientHandler(address)
	handler.Timeout = timeout
	handler.SlaveId = slaveID

	client := modbus.NewClient(handler)
	return client, nil
}
