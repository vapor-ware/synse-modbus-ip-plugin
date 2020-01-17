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
)
