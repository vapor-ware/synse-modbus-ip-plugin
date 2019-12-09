package testutils

import "errors"

// ErrFakeModbus is the error which the FakeModbus client returns when configured
// to return an error.
var ErrFakeModbus = errors.New("error for FakeModbus test client")

// FakeModbus implements the modbus.Client interface. It is used for testing.
type FakeModbus struct {
	withError bool
	response  []byte
}

// NewFakeModbusClient creates a new instance of FakeModbus which can be used for
// testing.
func NewFakeModbusClient() *FakeModbus {
	return &FakeModbus{
		response: []byte{},
	}
}

// WithError sets the FakeModbus client to return an error for any subsequent
// modbus call(s).
func (c *FakeModbus) WithError() *FakeModbus {
	c.withError = true
	return c
}

// WithResponse sets the response bytes which the FakeModbus client should return
// for any subsequent modbus call(s).
func (c *FakeModbus) WithResponse(resp []byte) *FakeModbus {
	c.response = resp
	return c
}

func (c *FakeModbus) ReadCoils(address, quantity uint16) (results []byte, err error) {
	if c.withError {
		return nil, ErrFakeModbus
	}
	return c.response, nil
}

func (c *FakeModbus) ReadDiscreteInputs(address, quantity uint16) (results []byte, err error) {
	if c.withError {
		return nil, ErrFakeModbus
	}
	return c.response, nil
}

func (c *FakeModbus) WriteSingleCoil(address, value uint16) (results []byte, err error) {
	if c.withError {
		return nil, ErrFakeModbus
	}
	return c.response, nil
}

func (c *FakeModbus) WriteMultipleCoils(address, quantity uint16, value []byte) (results []byte, err error) {
	if c.withError {
		return nil, ErrFakeModbus
	}
	return c.response, nil
}

func (c *FakeModbus) ReadInputRegisters(address, quantity uint16) (results []byte, err error) {
	if c.withError {
		return nil, ErrFakeModbus
	}
	return c.response, nil
}

func (c *FakeModbus) ReadHoldingRegisters(address, quantity uint16) (results []byte, err error) {
	if c.withError {
		return nil, ErrFakeModbus
	}
	return c.response, nil
}

func (c *FakeModbus) WriteSingleRegister(address, value uint16) (results []byte, err error) {
	if c.withError {
		return nil, ErrFakeModbus
	}
	return c.response, nil
}

func (c *FakeModbus) WriteMultipleRegisters(address, quantity uint16, value []byte) (results []byte, err error) {
	if c.withError {
		return nil, ErrFakeModbus
	}
	return c.response, nil
}

func (c *FakeModbus) ReadWriteMultipleRegisters(readAddress, readQuantity, writeAddress, writeQuantity uint16, value []byte) (results []byte, err error) {
	if c.withError {
		return nil, ErrFakeModbus
	}
	return c.response, nil
}

func (c *FakeModbus) MaskWriteRegister(address, andMask, orMask uint16) (results []byte, err error) {
	if c.withError {
		return nil, ErrFakeModbus
	}
	return c.response, nil
}

func (c *FakeModbus) ReadFIFOQueue(address uint16) (results []byte, err error) {
	if c.withError {
		return nil, ErrFakeModbus
	}
	return c.response, nil
}
