package main

import (
	log "github.com/sirupsen/logrus"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg"
	"github.com/vapor-ware/synse-sdk/sdk"
)

const (
	pluginName       = "modbus ip"
	pluginMaintainer = "vaporio"
	pluginDesc       = "A plugin for modbus over TCP/IP"
	pluginVcs        = "github.com/vapor-ware/synse-modbus-ip-plugin"
)

func main() {
	// Set the plugin metadata
	sdk.SetPluginInfo(
		pluginName,
		pluginMaintainer,
		pluginDesc,
		pluginVcs,
	)

	plugin := pkg.MakePlugin()

	// Run the plugin
	if err := plugin.Run(); err != nil {
		log.Fatal(err)
	}
}
