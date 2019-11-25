package devices

import (
	log "github.com/sirupsen/logrus"
)

// ByModbusConfig is a slice of *sdk.Device which implements the Sort
// interface. It is is sortable by the devices' configured modbus config,
// found in the Data field for the device.

type ByModbusConfig []*ModbusDevice

func (a ByModbusConfig) Len() int {
	return len(a)
}

func (a ByModbusConfig) Less(i, j int) bool {
	iHost, jHost := a[i].Config.Host, a[j].Config.Host
	if iHost < jHost {
		return true
	} else if iHost > jHost {
		return false
	}

	iPort, jPort := a[i].Config.Port, a[j].Config.Port
	if iPort < jPort {
		return true
	} else if iPort > jPort {
		return false
	}

	iAddr, jAddr := a[i].Config.Address, a[j].Config.Address
	if iAddr < jAddr {
		return true
	} else if iAddr > jAddr {
		return false
	}

	log.WithFields(log.Fields{
		"host":    iHost,
		"port":    iPort,
		"address": iAddr,
	}).Warning("duplicate modbus device configuration detected")
	return true
}

func (a ByModbusConfig) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
