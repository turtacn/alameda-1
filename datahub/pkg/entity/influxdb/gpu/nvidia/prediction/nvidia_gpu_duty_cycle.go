package prediction

import (
	"time"
)

type dutyCycleTag = string
type dutyCycleField = string

const (
	DutyCycleTime        dutyCycleTag = "time"
	DutyCycleHost        dutyCycleTag = "host"
	DutyCycleInstance    dutyCycleTag = "instance"
	DutyCycleJob         dutyCycleTag = "job"
	DutyCycleName        dutyCycleTag = "name"
	DutyCycleUuid        dutyCycleTag = "uuid"
	DutyCycleGranularity dutyCycleTag = "granularity"

	DutyCycleMinorNumber dutyCycleField = "minor_number"
	DutyCycleValue       dutyCycleField = "value"
)

type DutyCycleEntity struct {
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
