package devices

/*
import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/vapor-ware/synse-modbus-ip-plugin/pkg/config"
)

func TestByModbusConfig_Len(t *testing.T) {
	devs := ByModbusConfig([]*ModbusDevice{
		{Config: &config.ModbusConfig{Address: 1}},
		{Config: &config.ModbusConfig{Address: 2}},
		{Config: &config.ModbusConfig{Address: 3}},
	})
	assert.Equal(t, 3, devs.Len())
}

func TestByModbusConfig_Len_Empty(t *testing.T) {
	devs := ByModbusConfig([]*ModbusDevice{})
	assert.Equal(t, 0, devs.Len())
}

func TestSort_ByModbusConfig_Empty(t *testing.T) {
	devs := ByModbusConfig([]*ModbusDevice{})

	sort.Sort(devs)
	assert.Len(t, devs, 0)
}

func TestSort_ByModbusConfig_Single(t *testing.T) {
	devs := ByModbusConfig([]*ModbusDevice{
		{Config: &config.ModbusConfig{Address: 1}},
	})

	sort.Sort(devs)
	assert.Len(t, devs, 1)
	assert.Equal(t, uint16(1), devs[0].Config.Address)
}

func TestSort_ByModbusConfig_MultipleInOrder(t *testing.T) {
	devs := ByModbusConfig([]*ModbusDevice{
		// These configs hijack the Type field to provide identity in the test
		// assertions -- the type does not look like this in reality.
		{Config: &config.ModbusConfig{Host: "a", Port: 500, Address: 1, Type: "dev-1"}},
		{Config: &config.ModbusConfig{Host: "a", Port: 500, Address: 2, Type: "dev-2"}},
		{Config: &config.ModbusConfig{Host: "a", Port: 501, Address: 1, Type: "dev-3"}},
		{Config: &config.ModbusConfig{Host: "b", Port: 502, Address: 1, Type: "dev-4"}},
		{Config: &config.ModbusConfig{Host: "c", Port: 503, Address: 1, Type: "dev-5"}},
		{Config: &config.ModbusConfig{Host: "c", Port: 503, Address: 4, Type: "dev-6"}},
	})

	sort.Sort(devs)
	assert.Len(t, devs, 6)
	assert.Equal(t, "dev-1", devs[0].Config.Type)
	assert.Equal(t, "dev-2", devs[1].Config.Type)
	assert.Equal(t, "dev-3", devs[2].Config.Type)
	assert.Equal(t, "dev-4", devs[3].Config.Type)
	assert.Equal(t, "dev-5", devs[4].Config.Type)
	assert.Equal(t, "dev-6", devs[5].Config.Type)
}

func TestSort_ByModbusConfig_Duplicate(t *testing.T) {
	devs := ByModbusConfig([]*ModbusDevice{
		// These configs hijack the Type field to provide identity in the test
		// assertions -- the type does not look like this in reality.
		{Config: &config.ModbusConfig{Host: "a", Port: 500, Address: 1, Type: "dev-1"}},
		{Config: &config.ModbusConfig{Host: "a", Port: 500, Address: 1, Type: "dev-2"}},
		{Config: &config.ModbusConfig{Host: "a", Port: 500, Address: 1, Type: "dev-3"}},
	})

	sort.Sort(devs)
	assert.Len(t, devs, 3)
	assert.Equal(t, "dev-3", devs[0].Config.Type)
	assert.Equal(t, "dev-2", devs[1].Config.Type)
	assert.Equal(t, "dev-1", devs[2].Config.Type)
}

func TestSort_ByModbusConfig_MultipleOutOfOrder(t *testing.T) {
	devs := ByModbusConfig([]*ModbusDevice{
		// These configs hijack the Type field to provide identity in the test
		// assertions -- the type does not look like this in reality.
		{Config: &config.ModbusConfig{Host: "a", Port: 500, Address: 1, Type: "dev-1"}},
		{Config: &config.ModbusConfig{Host: "b", Port: 501, Address: 2, Type: "dev-2"}},
		{Config: &config.ModbusConfig{Host: "a", Port: 500, Address: 3, Type: "dev-3"}},
		{Config: &config.ModbusConfig{Host: "a", Port: 501, Address: 1, Type: "dev-4"}},
		{Config: &config.ModbusConfig{Host: "c", Port: 501, Address: 1, Type: "dev-5"}},
		{Config: &config.ModbusConfig{Host: "b", Port: 502, Address: 3, Type: "dev-6"}},
		{Config: &config.ModbusConfig{Host: "a", Port: 503, Address: 4, Type: "dev-7"}},
		{Config: &config.ModbusConfig{Host: "d", Port: 506, Address: 2, Type: "dev-8"}},
		{Config: &config.ModbusConfig{Host: "c", Port: 503, Address: 4, Type: "dev-9"}},
		{Config: &config.ModbusConfig{Host: "d", Port: 504, Address: 1, Type: "dev-10"}},
		{Config: &config.ModbusConfig{Host: "c", Port: 504, Address: 5, Type: "dev-11"}},
		{Config: &config.ModbusConfig{Host: "c", Port: 504, Address: 1, Type: "dev-12"}},
	})

	sort.Sort(devs)
	assert.Len(t, devs, 12)
	assert.Equal(t, "dev-1", devs[0].Config.Type)
	assert.Equal(t, "dev-3", devs[1].Config.Type)
	assert.Equal(t, "dev-4", devs[2].Config.Type)
	assert.Equal(t, "dev-7", devs[3].Config.Type)
	assert.Equal(t, "dev-2", devs[4].Config.Type)
	assert.Equal(t, "dev-6", devs[5].Config.Type)
	assert.Equal(t, "dev-5", devs[6].Config.Type)
	assert.Equal(t, "dev-9", devs[7].Config.Type)
	assert.Equal(t, "dev-12", devs[8].Config.Type)
	assert.Equal(t, "dev-11", devs[9].Config.Type)
	assert.Equal(t, "dev-10", devs[10].Config.Type)
	assert.Equal(t, "dev-8", devs[11].Config.Type)
}
*/
