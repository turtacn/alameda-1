package metrics

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"strconv"
	"time"
)

type memoryTotalBytesTag = string
type memoryTotalBytesField = string

const (
	MemoryTotalBytesTime     memoryTotalBytesTag = "time"
	MemoryTotalBytesHost     memoryTotalBytesTag = "host"
	MemoryTotalBytesInstance memoryTotalBytesTag = "instance"
	MemoryTotalBytesJob      memoryTotalBytesTag = "job"
	MemoryTotalBytesName     memoryTotalBytesTag = "name"
	MemoryTotalBytesUuid     memoryTotalBytesTag = "uuid"

	MemoryTotalBytesMinorNumber memoryTotalBytesField = "minor_number"
	MemoryTotalBytesValue       memoryTotalBytesField = "value"
)

type MemoryTotalBytesEntity struct {
	Time     time.Time
	Host     *string
	Instance *string
	Job      *string
	Name     *string
	Uuid     *string

	MinorNumber *string
	Value       *float64
}

func NewMemoryTotalBytesEntityFromMap(data map[string]string) MemoryTotalBytesEntity {
	tempTimestamp, _ := utils.ParseTime(data[MemoryTotalBytesTime])

	entity := MemoryTotalBytesEntity{
		Time: tempTimestamp,
	}

	// InfluxDB tags
	if valueStr, exist := data[MemoryTotalBytesHost]; exist {
		entity.Host = &valueStr
	}
	if valueStr, exist := data[MemoryTotalBytesInstance]; exist {
		entity.Instance = &valueStr
	}
	if valueStr, exist := data[MemoryTotalBytesJob]; exist {
		entity.Job = &valueStr
	}
	if valueStr, exist := data[MemoryTotalBytesName]; exist {
		entity.Name = &valueStr
	}
	if valueStr, exist := data[MemoryTotalBytesUuid]; exist {
		entity.Uuid = &valueStr
	}

	// InfluxDB fields
	if valueStr, exist := data[MemoryTotalBytesMinorNumber]; exist {
		entity.MinorNumber = &valueStr
	}
	if valueFloat, exist := data[MemoryTotalBytesValue]; exist {
		value, _ := strconv.ParseFloat(valueFloat, 64)
		entity.Value = &value
	}

	return entity
}
