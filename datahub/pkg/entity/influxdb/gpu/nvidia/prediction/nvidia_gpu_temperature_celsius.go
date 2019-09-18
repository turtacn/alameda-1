package prediction

import (
	"time"
)

type temperatureCelsiusTag = string
type temperatureCelsiusField = string

const (
	TemperatureCelsiusTime        temperatureCelsiusTag = "time"
	TemperatureCelsiusHost        temperatureCelsiusTag = "host"
	TemperatureCelsiusInstance    temperatureCelsiusTag = "instance"
	TemperatureCelsiusJob         temperatureCelsiusTag = "job"
	TemperatureCelsiusName        temperatureCelsiusTag = "name"
	TemperatureCelsiusUuid        temperatureCelsiusTag = "uuid"
	TemperatureCelsiusGranularity temperatureCelsiusTag = "granularity"

	TemperatureCelsiusMinorNumber temperatureCelsiusField = "minor_number"
	TemperatureCelsiusValue       temperatureCelsiusField = "value"
)

type TemperatureCelsiusEntity struct {
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
