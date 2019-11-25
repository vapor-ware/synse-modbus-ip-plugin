package testutils

import "errors"

var ErrFakeModbus = errors.New("error for FakeModbus test client")

type FakeModbus struct {
	withError bool
	response  []byte
}

func NewFakeModbusClient() *FakeModbus {
	return &FakeModbus{
		response: []byte{},
	}
}

func (c *FakeModbus) WithError() *FakeModbus {
	c.withError = true
	return c
}

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
