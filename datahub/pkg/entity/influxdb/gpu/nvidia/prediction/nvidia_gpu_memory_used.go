package prediction

import (
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
	MemoryUsedGranularity   memoryUsedBytesTag = "granularity"

	MemoryUsedBytesMinorNumber memoryUsedBytesField = "minor_number"
	MemoryUsedBytesValue       memoryUsedBytesField = "value"
)

type MemoryUsedBytesEntity struct {
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
