package pkg

import (
	log "github.com/Sirupsen/logrus"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/devices"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/outputs"
	"github.com/vapor-ware/synse-sdk/sdk"
)

// MakePlugin creates a new instance of the Synse Modbus TCP/IP Plugin.
func MakePlugin() *sdk.Plugin {
	plugin := sdk.NewPlugin()

	// Register the output types
	err := plugin.RegisterOutputTypes(
		&outputs.Current,
		&outputs.Power,
		&outputs.Voltage,
		&outputs.Frequency,
		&outputs.SItoKWhPower,
		&outputs.FanSpeedPercent,
		&outputs.TemperatureFTenths,
		&outputs.FlowGpm,
		&outputs.FlowGpmTenths,
		&outputs.Coil,
		&outputs.InWCThousanths,
		&outputs.PsiTenths,
	)
	if err != nil {
		log.Fatal(err)
	}

	// Register device handlers
	plugin.RegisterDeviceHandlers(
		&devices.CoilsHandler,
		&devices.HoldingRegisterHandler,
		&devices.InputRegisterHandler,
	)

	return plugin
}
