package pkg

import (
	"fmt"
)

// deviceIdentifier defines the Modbus-IP-specific way of uniquely identifying a device
// through its device configuration.
//
// FIXME - this is just a stub for framing up the plugin
func deviceIdentifier(data map[string]interface{}) string {
	return fmt.Sprint(data["id"])
}
