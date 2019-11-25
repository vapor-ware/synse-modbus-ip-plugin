package outputs

import "github.com/vapor-ware/synse-sdk/sdk/output"

var (

	// SItoKWhPower is the output type that converts a power reading in SI
	// (m2*kg/sec2(Joule)) to kilowatt hour (kWh). The conversion from kWh
	// to SI unit is 2.60E+06, so the inverse is used to convert from SI
	// to kWh 1/2.60E+06 ~= 2.77777778e-7
	SItoKWhPower = output.Output{
		Name:      "si-to-kwh.power",
		Precision: 5,
		Unit: &output.Unit{
			Name:   "kilowatt hour",
			Symbol: "kWh",
		},
	}

	// Power is the output type for power (W) readings.
	Power = output.Output{
		Name:      "power",
		Precision: 3,
		Unit: &output.Unit{
			Name:   "watt",
			Symbol: "W",
		},
	}

	// Microseconds in the output type for time readings in microseconds.
	Microseconds = output.Output{
		Name:      "microseconds",
		Precision: 6,
		Unit: &output.Unit{
			Name:   "microseconds",
			Symbol: "µs",
		},
	}

	// FanSpeedPercent is the output type for the VEM PLC fan.
	// This is a sliding window that is up to the PLC.
	// We do not get absolute rpm.
	FanSpeedPercent = output.Output{
		Name:      "fan_speed_percent",
		Precision: 3,
		Unit: &output.Unit{
			Name:   "percent",
			Symbol: "%",
		},
	}

	// FanSpeedPercentTenths is the output type for the VEM PLC fan in tenths.
	FanSpeedPercentTenths = output.Output{
		Name:      "fan_speed_percent_tenths",
		Precision: 3,
		Unit: &output.Unit{
			Name:   "percent",
			Symbol: "%",
		},
	}

	// Temperature is the output type for temperature readings.
	Temperature = output.Output{
		Name:      "temperature",
		Precision: 3,
		Unit: &output.Unit{
			Name:   "celsius",
			Symbol: "C",
		},
	}

	// FlowGpm is the output type for flow readings in gallons per minute. FUTURE: Metric / English.
	FlowGpm = output.Output{
		Name:      "flowGpm",
		Precision: 4,
		Unit: &output.Unit{
			Name:   "gallons per minute",
			Symbol: "gpm",
		},
	}

	// FlowGpmTenths is the output type for flow readings in tenths of gallons per minute. FUTURE: Metric / English.
	FlowGpmTenths = output.Output{
		Name:      "flowGpmTenths",
		Precision: 4,
		Unit: &output.Unit{
			Name:   "gallons per minute",
			Symbol: "gpm",
		},
	}

	// Coil is the output type for a coil.
	// VEM PLC coils are all active high, but this may vary with different devices.
	// Perhaps Synse should abstract this away and report all coils as active high
	// because Sysnse is a device level abstraction layer(?)
	Coil = output.Output{
		Name:      "switch",
		Precision: 1,
		Unit: &output.Unit{
			Name:   "",
			Symbol: "",
		},
	}

	// InWCThousanths is the output type for pressure readings measured in thousanths of inches of water column..
	InWCThousanths = output.Output{
		Name:      "InWCThousanths",
		Precision: 4,
		Unit: &output.Unit{
			Name:   "inches of water column",
			Symbol: "InWC",
		},
	}

	// PsiTenths is the output type for pressure readings measured in tenths of pounds per square inch..
	PsiTenths = output.Output{
		Name:      "psiTenths",
		Precision: 3,
		Unit: &output.Unit{
			Name:   "pounds per square inch",
			Symbol: "psi",
		},
	}

	// VoltSeconds is for flux.
	VoltSeconds = output.Output{
		Name:      "voltSeconds",
		Precision: 3,
		Unit: &output.Unit{
			Name:   "volt seconds",
			Symbol: "Vs",
		},
	}

	// FIXME: not clear that this belongs here, as this is Vapor-specific and not
	//   modbus related. We could define a position output in the SDK.

	// CarouselPosition is for the position of the Carousel, result is Wedge Id
	// facing the customer.
	CarouselPosition = output.Output{
		Name:      "carouselPosition",
		Precision: 3,
		Unit: &output.Unit{
			Name: "position",
		},
	}

	// StatusCode is the integer assosciated with a status response.
	StatusCode = output.Output{
		Name: "statusCode",
	}

	// StatusOutput is for an arbitrary string output which is meant to be the
	// string translation for status code.
	StatusOutput = output.Output{
		Name: "status",
	}
)
