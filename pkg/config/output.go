package config

// ModbusOutputData models the scheme for the supported config values
// of device output's Data field for the Modbus TCP/IP plugin.
type ModbusOutputData struct {
	// Address is the register address which holds the output reading.
	Address int

	// Width is the number of registers to read, starting from the `Address`.
	Width int

	// Type is the type of the data held in the registers. The supported
	// types are as follows:
	//
	//  "u32", "uint32":  unsigned 32-bit integer
	//  "u64", "uint64":  unsigned 64-bit integer
	//  "s32", "int32":   signed 32-bit integer
	//  "s64", "int64":   signed 64-bit integer
	//  "f32", "float32": 32-bit floating point number
	//  "f64", "float64": 64-bit floating point number
	Type string
}
