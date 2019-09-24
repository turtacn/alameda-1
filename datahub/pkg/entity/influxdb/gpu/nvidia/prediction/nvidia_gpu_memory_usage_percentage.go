package prediction

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"strconv"
	"time"
)

type memoryUsageTag = string
type memoryUsageField = string

const (
	MemoryUsageTime        memoryUsageTag = "time"
	MemoryUsageHost        memoryUsageTag = "host"
	MemoryUsageInstance    memoryUsageTag = "instance"
	MemoryUsageJob         memoryUsageTag = "job"
	MemoryUsageName        memoryUsageTag = "name"
	MemoryUsageUuid        memoryUsageTag = "uuid"
	MemoryUsageGranularity memoryUsageTag = "granularity"

	MemoryUsageMinorNumber memoryUsageField = "minor_number"
	MemoryUsageValue       memoryUsageField = "value"
)

type MemoryUsageEntity struct {
	Time        time.Time
	Host        *string
	Instance    *string
	Job         *string
	Name        *string
	Uuid        *string
	Granularity *string

	MinorNumber *string
	Value       *float64
}

func NewMemoryUsageEntityFromMap(data map[string]string) MemoryUsageEntity {
	tempTimestamp, _ := utils.ParseTime(data[MemoryUsageTime])

	entity := MemoryUsageEntity{
		Time: tempTimestamp,
	}

	// InfluxDB tags
	if valueStr, exist := data[MemoryUsageHost]; exist {
		entity.Host = &valueStr
	}
	if valueStr, exist := data[MemoryUsageInstance]; exist {
		entity.Instance = &valueStr
	}
	if valueStr, exist := data[MemoryUsageJob]; exist {
		entity.Job = &valueStr
	}
	if valueStr, exist := data[MemoryUsageName]; exist {
		entity.Name = &valueStr
	}
	if valueStr, exist := data[MemoryUsageUuid]; exist {
		entity.Uuid = &valueStr
	}
	if valueStr, exist := data[MemoryUsageGranularity]; exist {
		entity.Granularity = &valueStr
	}

	// InfluxDB fields
	if valueStr, exist := data[MemoryUsageMinorNumber]; exist {
		entity.MinorNumber = &valueStr
	}
	if valueFloat, exist := data[MemoryUsageValue]; exist {
		value, _ := strconv.ParseFloat(valueFloat, 64)
		entity.Value = &value
	}

	return entity
}
