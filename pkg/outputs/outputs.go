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

	// FanSpeedPercent is the output type for the VEM PLC fan.
	// This is a sliding window that is up to the PLC.
	// We do not get absolute rpm.
	FanSpeedPercent = sdk.OutputType{
		Name:      "fan_speed_percent",
		Precision: 3,
		Unit: sdk.Unit{
			Name:   "percent",
			Symbol: "%",
		},
	}

	// Temperature is the output type for temperature (C) readings.
	Temperature = sdk.OutputType{
		Name:      "temperature",
		Precision: 3,
		Unit: sdk.Unit{
			Name:   "celsius",
			Symbol: "C",
		},
	}

	// FlowGpm is the output type for flow readings in gallons per minute. FUTURE: Metric / English.
	FlowGpm = sdk.OutputType{
		Name:      "flowGpm",
		Precision: 4,
		Unit: sdk.Unit{
			Name:   "gallons per minute",
			Symbol: "gpm",
		},
	}

	// Coil is the output type for a coil.
	// VEM PLC coils are all active high, but this may vary with different devices.
	// Perhaps Synse should abstract this away and report all coils as active high
	// because Sysnse is a device level abstraction layer(?)
	Coil = sdk.OutputType{
		Name:      "boolean",
		Precision: 1,
		Unit: sdk.Unit{
			Name:   "",
			Symbol: "",
		},
	}

	// InWC is the output type for pressure readings measured in inches of water column..
	InWC = sdk.OutputType{
		Name:      "InWC",
		Precision: 4,
		Unit: sdk.Unit{
			Name:   "inches of water column",
			Symbol: "InWC",
		},
	}

	// Psi is the output type for pressure readings measured in pounds per square inch..
	Psi = sdk.OutputType{
		Name:      "psi",
		Precision: 3,
		Unit: sdk.Unit{
			Name:   "pounds per square inch",
			Symbol: "psi",
		},
	}
)
