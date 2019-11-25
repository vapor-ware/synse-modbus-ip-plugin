package devices

import (
	"testing"

	"github.com/stretchr/testify/suite"
	"github.com/vapor-ware/synse-sdk/sdk"
)

type NewModbusClientTestSuite struct {
	suite.Suite
}

func TestModbusClientTestSuite(t *testing.T) {
	suite.Run(t, new(NewModbusClientTestSuite))
}

func (suite *NewModbusClientTestSuite) TestOK() {
	dev := sdk.Device{
		Data: map[string]interface{}{
			"host": "localhost",
			"port": 7777,
		},
	}

	client, err := NewModbusClient(&dev)
	suite.NoError(err)
	suite.NotNil(client)
}

func (suite *NewModbusClientTestSuite) TestDecodeError() {
	dev := sdk.Device{
		Data: map[string]interface{}{
			"host": "localhost",
			"port": "not-an-int",
		},
	}

	client, err := NewModbusClient(&dev)
	suite.Error(err)
	suite.Nil(client)
}

func (suite *NewModbusClientTestSuite) TestNewClientError_FailOnError() {
	dev := sdk.Device{
		Data: map[string]interface{}{
			// missing required field: host
			"port":        7777,
			"failOnError": true,
		},
	}

	client, err := NewModbusClient(&dev)
	suite.Error(err)
	suite.Nil(client)
}

func (suite *NewModbusClientTestSuite) TestNewClientError_NoFailOnError() {
	dev := sdk.Device{
		Data: map[string]interface{}{
			// missing required field: host
			"port":        7777,
			"failOnError": false,
		},
	}

	client, err := NewModbusClient(&dev)
	suite.NoError(err)
	suite.Nil(client)
}
