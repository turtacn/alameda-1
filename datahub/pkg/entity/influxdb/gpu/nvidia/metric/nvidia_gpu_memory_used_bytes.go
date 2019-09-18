package metric

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"strconv"
	"time"
)

type memoryUsedBytesTag = string
type memoryUsedBytesField = string

const (
	MemoryUsedBytesTime     memoryUsedBytesTag = "time"
	MemoryUsedBytesHost     memoryUsedBytesTag = "host"
	MemoryUsedBytesInstance memoryUsedBytesTag = "instance"
	MemoryUsedBytesJob      memoryUsedBytesTag = "job"
	MemoryUsedBytesName     memoryUsedBytesTag = "name"
	MemoryUsedBytesUuid     memoryUsedBytesTag = "uuid"

	MemoryUsedBytesMinorNumber memoryUsedBytesField = "minor_number"
	MemoryUsedBytesValue       memoryUsedBytesField = "value"
)

type MemoryUsedBytesEntity struct {
	Time     time.Time
	Host     *string
	Instance *string
	Job      *string
	Name     *string
	Uuid     *string

	MinorNumber *string
	Value       *float64
}

func NewMemoryUsedBytesEntityFromMap(data map[string]string) MemoryUsedBytesEntity {
	tempTimestamp, _ := utils.ParseTime(data[MemoryUsedBytesTime])

	entity := MemoryUsedBytesEntity{
		Time: tempTimestamp,
	}

	// InfluxDB tags
	if valueStr, exist := data[MemoryUsedBytesHost]; exist {
		entity.Host = &valueStr
	}
	if valueStr, exist := data[MemoryUsedBytesInstance]; exist {
		entity.Instance = &valueStr
	}
	if valueStr, exist := data[MemoryUsedBytesJob]; exist {
		entity.Job = &valueStr
	}
	if valueStr, exist := data[MemoryUsedBytesName]; exist {
		entity.Name = &valueStr
	}
	if valueStr, exist := data[MemoryUsedBytesUuid]; exist {
		entity.Uuid = &valueStr
	}

	// InfluxDB fields
	if valueStr, exist := data[MemoryUsedBytesMinorNumber]; exist {
		entity.MinorNumber = &valueStr
	}
	if valueFloat, exist := data[MemoryUsedBytesValue]; exist {
		value, _ := strconv.ParseFloat(valueFloat, 64)
		entity.Value = &value
	}

	return entity
}
