package pkg

import (
	"log"

	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/devices"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/outputs"
	"github.com/vapor-ware/synse-sdk/sdk"
)

// MakePlugin creates a new instance of the Synse Modbus-IP Plugin.
func MakePlugin() *sdk.Plugin {
	plugin := sdk.NewPlugin(
		sdk.CustomDeviceIdentifier(deviceIdentifier),
	)

	// Register the output types
	err := plugin.RegisterOutputTypes(
		&outputs.Current,
		&outputs.Power,
		&outputs.Voltage,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register device handlers
	plugin.RegisterDeviceHandlers(
		&devices.EG4115PowerMeter,
	)

	return plugin
}
