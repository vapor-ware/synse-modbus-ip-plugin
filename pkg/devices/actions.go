package devices

import (
	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-sdk/sdk"
)

/*
// DeviceManagers maps device handler to their corresponding device managers. Devices
// are associated to handlers via their device configuration.
//
// This mapping provides the same information about associated devices as would the
// `devices` parameter from a DeviceHandler's BulkRead function. The SDK tracks which
// devices belong to which handler, just as this map does albeit with greater data
// complexity. As such, the Devices to read should come from a manager, not from the
// Device slice provided by the SDK.
var DeviceManagers = map[string][]*ModbusDeviceManager{}

// clearDeviceManagers is a utility function used by tests which is used to clean up the
// DeviceManagers map between test runs. It should not be used outside of testing.
func clearDeviceManagers() {
	DeviceManagers = map[string][]*ModbusDeviceManager{}
}
*/

// TODO: This function below is weird because it doesn't return anything.
// It just contructs some thinggy internally and hides it. (DeviceManglers)

// LoadModbusDevices is an SDK DeviceAction which loads all registered devices into
// ModbusDevices, a higher-level wrapper which aggregates devices based on their modbus
// config. This allows for bulk actions across contiguous register blocks.
var LoadModbusDevices = sdk.DeviceAction{
	Name: "load-modbus-devices",
	Filter: map[string][]string{
		"type": {"*"}, // All devices
	},
	Action: func(p *sdk.Plugin, d *sdk.Device) (err error) {
		// Create a new ModbusDevice wrapper for the given device. This will parse
		// the Device's Data field into a struct for easier access.
		dev, err := NewModbusDevice(d)
		if err != nil {
			log.WithError(err).Error("failed to create new ModbusDevice wrapper")
			return err
		}

		// Get the ModbusDeviceManager for the device's handler. Check if a handler
		// manager exists for the device. If one does not, create a new one.
		managers, found := DeviceManagers[d.Handler]
		if !found {
			log.WithFields(log.Fields{
				"handler": d.Handler,
				"host":    dev.Config.Host,
				"port":    dev.Config.Port,
			}).Debug("no device managers found for handler, creating new manager")
			manager, err := NewModbusDeviceManager(dev)
			if err != nil {
				return err
			}
			manager.Sort()
			DeviceManagers[d.Handler] = []*ModbusDeviceManager{manager}
		} else {
			log.WithFields(log.Fields{
				"handler": d.Handler,
			}).Debug("found manager(s) for handler - searching for match")
			var matched bool
			for _, m := range managers {
				if m.MatchesDevice(dev) {
					log.Debug("found match - adding device to existing manager")
					m.AddDevice(dev)
					m.Sort()
					matched = true
					break
				}
			}

			if !matched {
				log.Debug("no match found, creating new manager")
				manager, err := NewModbusDeviceManager(dev)
				if err != nil {
					return err
				}
				manager.Sort()
				DeviceManagers[d.Handler] = append(DeviceManagers[d.Handler], manager)
			}
		}

		return nil
	},
}
