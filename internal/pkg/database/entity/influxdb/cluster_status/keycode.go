package cluster_status

import (
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

type keycodeTag = string
type keycodeField = string

const (
	// Node is node measurement
	KeycodeMeasurement influxdb.Measurement = "keycode"
)

const (
	Keycode keycodeTag = "keycode"

	KeycodeStatus          keycodeField = "status"
	KeycodeType            keycodeField = "type"
	KeycodeState           keycodeField = "state"
	KeycodeRegistered      keycodeField = "registered"
	KeycodeExpireTimestamp keycodeField = "expire_timestamp"
	KeycodeRawdata         keycodeField = "rawdata"
)

var (
	// List of tags of keycode measurement
	KeycodeTags = []keycodeTag{
		Keycode,
	}
	// List of fields of keycode measurement
	KeycodeFields = []keycodeField{
		KeycodeStatus,
		KeycodeType,
		KeycodeState,
		KeycodeRegistered,
		KeycodeExpireTimestamp,
		KeycodeRawdata,
	}
)
