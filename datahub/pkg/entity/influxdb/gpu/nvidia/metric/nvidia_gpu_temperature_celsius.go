package metric

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"strconv"
	"time"
)

type temperatureCelsiusTag = string
type temperatureCelsiusField = string

const (
	TemperatureCelsiusTime     temperatureCelsiusTag = "time"
	TemperatureCelsiusHost     temperatureCelsiusTag = "host"
	TemperatureCelsiusInstance temperatureCelsiusTag = "instance"
	TemperatureCelsiusJob      temperatureCelsiusTag = "job"
	TemperatureCelsiusName     temperatureCelsiusTag = "name"
	TemperatureCelsiusUuid     temperatureCelsiusTag = "uuid"

	TemperatureCelsiusMinorNumber temperatureCelsiusField = "minor_number"
	TemperatureCelsiusValue       temperatureCelsiusField = "value"
)

type TemperatureCelsiusEntity struct {
	Time     time.Time
	Host     *string
	Instance *string
	Job      *string
	Name     *string
	Uuid     *string

	MinorNumber *string
	Value       *float64
}

func NewTemperatureCelsiusEntityFromMap(data map[string]string) TemperatureCelsiusEntity {
	tempTimestamp, _ := utils.ParseTime(data[TemperatureCelsiusTime])

	entity := TemperatureCelsiusEntity{
		Time: tempTimestamp,
	}

	// InfluxDB tags
	if valueStr, exist := data[TemperatureCelsiusHost]; exist {
		entity.Host = &valueStr
	}
	if valueStr, exist := data[TemperatureCelsiusInstance]; exist {
		entity.Instance = &valueStr
	}
	if valueStr, exist := data[TemperatureCelsiusJob]; exist {
		entity.Job = &valueStr
	}
	if valueStr, exist := data[TemperatureCelsiusName]; exist {
		entity.Name = &valueStr
	}
	if valueStr, exist := data[TemperatureCelsiusUuid]; exist {
		entity.Uuid = &valueStr
	}

	// InfluxDB fields
	if valueStr, exist := data[TemperatureCelsiusMinorNumber]; exist {
		entity.MinorNumber = &valueStr
	}
	if valueFloat, exist := data[TemperatureCelsiusValue]; exist {
		value, _ := strconv.ParseFloat(valueFloat, 64)
		entity.Value = &value
	}

	return entity
}
