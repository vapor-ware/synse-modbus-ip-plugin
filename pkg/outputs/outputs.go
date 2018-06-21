package outputs

import "github.com/vapor-ware/synse-sdk/sdk"

var (
	// Current is the output type for current (amp) readings.
	Current = sdk.OutputType{
		Name:      "current",
		Precision: 3,
		Unit: sdk.Unit{
			Name:   "ampere",
			Symbol: "A",
		},
	}

	// Voltage is the output type for voltage (volt) readings.
	Voltage = sdk.OutputType{
		Name:      "voltage",
		Precision: 3,
		Unit: sdk.Unit{
			Name:   "volt",
			Symbol: "V",
		},
	}

	// Power is the output type for power (kWh) readings.
	Power = sdk.OutputType{
		Name:      "power",
		Precision: 3,
		Unit: sdk.Unit{
			Name:   "kilowatt hour",
			Symbol: "kWh",
		},
	}
)
