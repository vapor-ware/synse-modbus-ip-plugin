package outputs

import "github.com/vapor-ware/synse-sdk/sdk/output"

var (

	// GallonsPerMin is the output type for volumetric flow rate, measured in gallons per minute.
	GallonsPerMin = output.Output{
		Name:      "gallonsPerMin",
		Type:      "flow",
		Precision: 4,
		Unit: &output.Unit{
			Name:   "gallons per minute",
			Symbol: "gpm",
		},
	}

	// InchesWaterColumn is the output type for pressure readings, measures in inches of water
	// column.
	InchesWaterColumn = output.Output{
		Name:      "inchesWaterColumn",
		Type:      "pressure",
		Precision: 8,
		Unit: &output.Unit{
			Name:   "inches of water column",
			Symbol: "inch wc",
		},
	}

	// MacAddressWide is a mac address with a full byte for each octet.
	MacAddressWide = output.Output{
		Name:      "macAddressWide",
		Type:      "macaddresswide",
		Precision: 6,
	}

	// PowerFactor is the ratio of real power to apparent power.
	PowerFactor = output.Output{
		Name:      "powerFactor",
		Type:      "powerFactor",
		Precision: 4,
	}

	// VoltAmp is apparent power.
	VoltAmp = output.Output{
		Name:      "voltamp",
		Type:      "voltamp",
		Precision: 4,
		Unit: &output.Unit{
			Name:   "volt amp",
			Symbol: "va",
		},
	}

	// VAR is reactive power.
	VAR = output.Output{
		Name:      "var",
		Type:      "var",
		Precision: 4,
		Unit: &output.Unit{
			Name:   "volt amp reactive",
			Symbol: "var",
		},
	}

	// WattHour is power over time.
	WattHour = output.Output{
		Name:      "watt-hour",
		Type:      "watt-hour",
		Precision: 4,
		Unit: &output.Unit{
			Name:   "watt hour",
			Symbol: "Wh",
		},
	}

	// Bytes are arbitrary bytes.
	Bytes = output.Output{
		Name: "bytes",
		Type: "bytes",
	}
)
