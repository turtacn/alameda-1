package metrics

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"strconv"
	"time"
)

type numDevicesTag = string
type numDevicesField = string

const (
	NumDevicesTime     numDevicesTag = "time"
	NumDevicesHost     numDevicesTag = "host"
	NumDevicesInstance numDevicesTag = "instance"
	NumDevicesJob      numDevicesTag = "job"

	NumDevicesValue numDevicesField = "value"
)

type NumDevicesEntity struct {
	Time     time.Time
	Host     *string
	Instance *string
	Job      *string

	Value *float64
}

func NewNumDevicesEntityFromMap(data map[string]string) NumDevicesEntity {
	tempTimestamp, _ := utils.ParseTime(data[NumDevicesTime])

	entity := NumDevicesEntity{
		Time: tempTimestamp,
	}

	// InfluxDB tags
	if valueStr, exist := data[NumDevicesHost]; exist {
		entity.Host = &valueStr
	}
	if valueStr, exist := data[NumDevicesInstance]; exist {
		entity.Instance = &valueStr
	}
	if valueStr, exist := data[NumDevicesJob]; exist {
		entity.Job = &valueStr
	}

	// InfluxDB fields
	if valueFloat, exist := data[NumDevicesValue]; exist {
		value, _ := strconv.ParseFloat(valueFloat, 64)
		entity.Value = &value
	}

	return entity
}
