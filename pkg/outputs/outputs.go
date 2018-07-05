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

	// SItoKWhPower is the output type that converts a power reading in SI
	// (m2*kg/sec2(Joule)) to kilowatt hour (kWh). The conversion from kWh
	// to SI unit is 2.60E+06, so the inverse is used to convert from SI
	// to kWh 1/2.60E+06 ~= 2.77777778e-7
	SItoKWhPower = sdk.OutputType{
		Name:          "si-to-kwh.power",
		Precision:     5,
		ScalingFactor: "2.77777778e-7",
		Unit: sdk.Unit{
			Name:   "kilowatt hour",
			Symbol: "kWh",
		},
	}

	// Power is the output type for power (W) readings.
	Power = sdk.OutputType{
		Name:      "power",
		Precision: 3,
		Unit: sdk.Unit{
			Name:   "watt",
			Symbol: "W",
		},
	}

	// Frequency is the output type for frequency (Hz) readings.
	Frequency = sdk.OutputType{
		Name:      "frequency",
		Precision: 3,
		Unit: sdk.Unit{
			Name:   "hertz",
			Symbol: "Hz",
		},
	}
)
