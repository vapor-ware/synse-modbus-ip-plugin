package config

import (
	"testing"
	//"time"
	"github.com/stretchr/testify/assert"
	sdkConfig "github.com/vapor-ware/synse-sdk/sdk/config"
	"github.com/vapor-ware/synse-sdk/sdk/policy"
)

// TestLargeDeviceConfig is testing a production busway config that was panicing.
/*
time="2021-01-06T23:21:37.088Z" level=info msg="[config] reading config file" file=/etc/synse/plugin/config/device/config.yml
panic: reflect: call of reflect.flag.mustBeExported on zero Value

goroutine 1 [running]:
reflect.flag.mustBeExportedSlow(0x0)
	/usr/local/go/src/reflect/value.go:222 +0xad
reflect.flag.mustBeExported(...)
	/usr/local/go/src/reflect/value.go:216
reflect.Value.Set(0xa6d060, 0xc000d5f610, 0x194, 0x0, 0x0, 0x0)
	/usr/local/go/src/reflect/value.go:1527 +0x56
github.com/vapor-ware/synse-sdk/sdk/utils.redactRecursive(0xa6d060, 0xc000d5f610, 0x194, 0xa6d060, 0xc000d5f600, 0x94)
	/home/mhink/go/pkg/mod/github.com/vapor-ware/synse-sdk@v0.1.0-alpha.0.20200724155421-64f57718083f/sdk/utils/redact.go:73 +0xb9a
github.com/vapor-ware/synse-sdk/sdk/utils.redactRecursive(0xa7c980, 0xc0000101e0, 0x195, 0xa7c980, 0xc000113290, 0x15)
	/home/mhink/go/pkg/mod/github.com/vapor-ware/synse-sdk@v0.1.0-alpha.0.20200724155421-64f57718083f/sdk/utils/redact.go:127 +0x77d
github.com/vapor-ware/synse-sdk/sdk/utils.redactRecursive(0xa6d060, 0xc000d5f440, 0x194, 0xa6d060, 0xc000d5f430, 0x94)
	/home/mhink/go/pkg/mod/github.com/vapor-ware/synse-sdk@v0.1.0-alpha.0.20200724155421-64f57718083f/sdk/utils/redact.go:76 +0xb1b
github.com/vapor-ware/synse-sdk/sdk/utils.redactRecursive(0xa7c980, 0xc0000101d8, 0x195, 0xa7c980, 0xc000112f30, 0x15)
	/home/mhink/go/pkg/mod/github.com/vapor-ware/synse-sdk@v0.1.0-alpha.0.20200724155421-64f57718083f/sdk/utils/redact.go:127 +0x77d
github.com/vapor-ware/synse-sdk/sdk/utils.redactRecursive(0xa6d060, 0xc000997510, 0x194, 0xa6d060, 0xc000996010, 0x194)
	/home/mhink/go/pkg/mod/github.com/vapor-ware/synse-sdk@v0.1.0-alpha.0.20200724155421-64f57718083f/sdk/utils/redact.go:76 +0xb1b
github.com/vapor-ware/synse-sdk/sdk/utils.redactRecursive(0xa39ee0, 0xc000d4ce60, 0x197, 0xa39ee0, 0xc000d4ce40, 0x97)
	/home/mhink/go/pkg/mod/github.com/vapor-ware/synse-sdk@v0.1.0-alpha.0.20200724155421-64f57718083f/sdk/utils/redact.go:84 +0xd91
github.com/vapor-ware/synse-sdk/sdk/utils.redactRecursive(0xa6d060, 0xc000d5e990, 0x194, 0xa6d060, 0xc000d5e980, 0x94)
	/home/mhink/go/pkg/mod/github.com/vapor-ware/synse-sdk@v0.1.0-alpha.0.20200724155421-64f57718083f/sdk/utils/redact.go:76 +0xb1b
github.com/vapor-ware/synse-sdk/sdk/utils.redactRecursive(0xa7ec00, 0xc000010178, 0x195, 0xa7ec00, 0xc0000f51a0, 0x15)
	/home/mhink/go/pkg/mod/github.com/vapor-ware/synse-sdk@v0.1.0-alpha.0.20200724155421-64f57718083f/sdk/utils/redact.go:127 +0x77d
github.com/vapor-ware/synse-sdk/sdk/utils.RedactPasswords(0xa7ec00, 0xc0000f51a0, 0xb4b841, 0x4)
	/home/mhink/go/pkg/mod/github.com/vapor-ware/synse-sdk@v0.1.0-alpha.0.20200724155421-64f57718083f/sdk/utils/redact.go:46 +0x14d
github.com/vapor-ware/synse-sdk/sdk/config.(*Loader).read(0xc0000c0370, 0xb4f35f, 0x8, 0x0, 0x0)
	/home/mhink/go/pkg/mod/github.com/vapor-ware/synse-sdk@v0.1.0-alpha.0.20200724155421-64f57718083f/sdk/config/config.go:426 +0x36f
github.com/vapor-ware/synse-sdk/sdk/config.(*Loader).Load(0xc0000c0370, 0xb4f35f, 0x8, 0x2, 0xc0001f1dc0)
	/home/mhink/go/pkg/mod/github.com/vapor-ware/synse-sdk@v0.1.0-alpha.0.20200724155421-64f57718083f/sdk/config/config.go:159 +0x42f
github.com/vapor-ware/synse-sdk/sdk.(*deviceManager).loadConfig(0xc0000f73b0, 0xc000000004, 0xc0001f1e18)
	/home/mhink/go/pkg/mod/github.com/vapor-ware/synse-sdk@v0.1.0-alpha.0.20200724155421-64f57718083f/sdk/device_manager.go:548 +0x199
github.com/vapor-ware/synse-sdk/sdk.(*deviceManager).init(0xc0000f73b0, 0xc000000004, 0xc0001f1e60)
	/home/mhink/go/pkg/mod/github.com/vapor-ware/synse-sdk@v0.1.0-alpha.0.20200724155421-64f57718083f/sdk/device_manager.go:111 +0x83
github.com/vapor-ware/synse-sdk/sdk.(*Plugin).initialize(0xc0000b43f0, 0xa4cfa0, 0x10eac00)
	/home/mhink/go/pkg/mod/github.com/vapor-ware/synse-sdk@v0.1.0-alpha.0.20200724155421-64f57718083f/sdk/plugin.go:311 +0x87
github.com/vapor-ware/synse-sdk/sdk.(*Plugin).Run(0xc0000b43f0, 0x0, 0xc0001f1f48)
	/home/mhink/go/pkg/mod/github.com/vapor-ware/synse-sdk@v0.1.0-alpha.0.20200724155421-64f57718083f/sdk/plugin.go:189 +0x32
main.main()
	/home/mhink/go/src/github.com/vapor-ware/synse-modbus-ip-plugin/main.go:28 +0xd3
*/
func TestLargeDeviceConfig(t *testing.T) {
	// Create a loader and load the device config.
	loader := sdkConfig.NewYamlLoader("some-loader")
	loader.AddSearchPaths("./testdata")
	err := loader.Load(policy.Required)
	assert.NoError(t, err)

	// Serialize to Devices.
	//devices := &sdk.Devices{}
	devices := &sdkConfig.Devices{}
	err = loader.Scan(devices)
	assert.NoError(t, err)
	assert.Equal(t, 3, devices.Version)
	t.Logf("--- DEVICES ---")
	t.Logf("len(devices.Devices): %v", len(devices.Devices))
	for i := 0; i < len(devices.Devices); i++ {
		t.Logf("devices.Devices[%v]: %#v", i, devices.Devices[i])
	}
	//t.Logf("%#v", devices)
}
