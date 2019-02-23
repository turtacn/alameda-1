package utils

import (
	"encoding/json"
)

// InterfaceToString encodes interface to string
func InterfaceToString(data interface{}) string {
	if configBin, err := json.MarshalIndent(data, "", "  "); err != nil {
		return ""
	} else {
		return string(configBin)
	}
}
