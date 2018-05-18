package devices

import (
	"fmt"

	"github.com/vapor-ware/synse-modbus-ip-plugin/protocol"
	"github.com/vapor-ware/synse-sdk/sdk"
)

/*
NOTE:

For the first rev of this plugin, we have the device handler defined
here explicitly. Once the SDK is updated to better support generalized
plugin structures, this will be updated and specific device handlers
will be replaced with a generalized handler.
*/

// EG4115PowerMeter is the handler for the eGauge 4115 Power Meter device.
var EG4115PowerMeter = sdk.DeviceHandler{
	Type:  "power",
	Model: "EG4115",

	Read: readEG4115PowerMeter,

	// The EG4115 Power Meter does not support writing.
	Write: nil,
}

func readEG4115PowerMeter(device *sdk.Device) ([]*sdk.Reading, error) {

	client, err := protocol.NewClient(device.Data)
	if err != nil {
		return nil, err
	}

	// FIXME (etd) - for now, the handling for the EG4115 power meter device is
	// going to be completely hardcoded. this plugin is not generalizable yet, as
	// there are many changes that need to happen in the next rev of the SDK first.
	// This gets things working for the power meter, but thats just about it.
	results, err := client.ReadInputRegisters(500, 8)
	if err != nil {
		return nil, err
	}

	l1RMS := protocol.Float32FromBytes(results[0:4])
	l2RMS := protocol.Float32FromBytes(results[4:8])
	l1l2RMS := protocol.Float32FromBytes(results[12:16])

	readings := []*sdk.Reading{
		sdk.NewReading("voltage", fmt.Sprintf("%f", l1RMS)),
		sdk.NewReading("voltage", fmt.Sprintf("%f", l2RMS)),
		sdk.NewReading("voltage", fmt.Sprintf("%f", l1l2RMS)),
	}
	return readings, nil
}
