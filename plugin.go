package main

import (
	"log"

	"github.com/vapor-ware/synse-sdk/sdk"

	"github.com/vapor-ware/synse-modbus-ip-plugin/devices"
)

// Build time variables for setting the version info of a Plugin.
var (
	BuildDate     string
	GitCommit     string
	GitTag        string
	GoVersion     string
	VersionString string
)

// DeviceIdentifier defines the Modbus-IP-specific way of uniquely identifying a device
// through its device configuration.
//
// FIXME - this is just a stub for framing up the plugin
func DeviceIdentifier(data map[string]string) string {
	return data["id"]
}

func main() {

	handlers, err := sdk.NewHandlers(DeviceIdentifier, nil)
	if err != nil {
		log.Fatal(err)
	}

	plugin, err := sdk.NewPlugin(handlers, nil)
	if err != nil {
		log.Fatal(err)
	}

	plugin.RegisterDeviceHandlers(
		&devices.EG4115PowerMeter,
	)

	// Set build-time version info.
	plugin.SetVersion(sdk.VersionInfo{
		BuildDate:     BuildDate,
		GitCommit:     GitCommit,
		GitTag:        GitTag,
		GoVersion:     GoVersion,
		VersionString: VersionString,
	})

	// Run the plugin.
	err = plugin.Run()
	if err != nil {
		log.Fatal(err)
	}
}
