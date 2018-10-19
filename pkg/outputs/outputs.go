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

	// TemperatureFTenths is the output type for temperature readings in tenths (multiply raw reading by .1)"
	TemperatureFTenths = sdk.OutputType{
		Name:          "temperatureFTenths",
		Precision:     3,
		ScalingFactor: ".1", // Raw reading for VEM PLC is tenths of degrees F.
		Unit: sdk.Unit{
			Name:   "fahrenheit",
			Symbol: "F",
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

	// FlowGpmTenths is the output type for flow readings in tenths of gallons per minute. FUTURE: Metric / English.
	FlowGpmTenths = sdk.OutputType{
		Name:          "flowGpmTenths",
		ScalingFactor: ".1",
		Precision:     4,
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
		Name:      "switch",
		Precision: 1,
		Unit: sdk.Unit{
			Name:   "",
			Symbol: "",
		},
	}

	// InWCThousanths is the output type for pressure readings measured in thousanths of inches of water column..
	InWCThousanths = sdk.OutputType{
		Name:          "InWCThousanths",
		ScalingFactor: ".001",
		Precision:     4,
		Unit: sdk.Unit{
			Name:   "inches of water column",
			Symbol: "InWC",
		},
	}

	// PsiTenths is the output type for pressure readings measured in tenths of pounds per square inch..
	PsiTenths = sdk.OutputType{
		Name:          "psiTenths",
		ScalingFactor: ".1",
		Precision:     3,
		Unit: sdk.Unit{
			Name:   "pounds per square inch",
			Symbol: "psi",
		},
	}
)
