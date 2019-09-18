package metric

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"strconv"
	"time"
)

type dutyCycleTag = string
type dutyCycleField = string

const (
	DutyCycleTime     dutyCycleTag = "time"
	DutyCycleHost     dutyCycleTag = "host"
	DutyCycleInstance dutyCycleTag = "instance"
	DutyCycleJob      dutyCycleTag = "job"
	DutyCycleName     dutyCycleTag = "name"
	DutyCycleUuid     dutyCycleTag = "uuid"

	DutyCycleMinorNumber dutyCycleField = "minor_number"
	DutyCycleValue       dutyCycleField = "value"
)

type DutyCycleEntity struct {
	Time     time.Time
	Host     *string
	Instance *string
	Job      *string
	Name     *string
	Uuid     *string

	MinorNumber *string
	Value       *float64
}

func NewDutyCycleEntityFromMap(data map[string]string) DutyCycleEntity {
	tempTimestamp, _ := utils.ParseTime(data[DutyCycleTime])

	entity := DutyCycleEntity{
		Time: tempTimestamp,
	}

	// InfluxDB tags
	if valueStr, exist := data[DutyCycleHost]; exist {
		entity.Host = &valueStr
	}
	if valueStr, exist := data[DutyCycleInstance]; exist {
		entity.Instance = &valueStr
	}
	if valueStr, exist := data[DutyCycleJob]; exist {
		entity.Job = &valueStr
	}
	if valueStr, exist := data[DutyCycleName]; exist {
		entity.Name = &valueStr
	}
	if valueStr, exist := data[DutyCycleUuid]; exist {
		entity.Uuid = &valueStr
	}

	// InfluxDB fields
	if valueStr, exist := data[DutyCycleMinorNumber]; exist {
		entity.MinorNumber = &valueStr
	}
	if valueFloat, exist := data[DutyCycleValue]; exist {
		value, _ := strconv.ParseFloat(valueFloat, 64)
		entity.Value = &value
	}

	return entity
}
