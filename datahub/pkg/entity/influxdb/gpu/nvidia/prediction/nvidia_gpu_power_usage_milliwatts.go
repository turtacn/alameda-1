package prediction

import (
	"time"
)

type powerUsageMilliWattsTag = string
type powerUsageMilliWattsField = string

const (
	PowerUsageMilliWattsTime        powerUsageMilliWattsTag = "time"
	PowerUsageMilliWattsHost        powerUsageMilliWattsTag = "host"
	PowerUsageMilliWattsInstance    powerUsageMilliWattsTag = "instance"
	PowerUsageMilliWattsJob         powerUsageMilliWattsTag = "job"
	PowerUsageMilliWattsName        powerUsageMilliWattsTag = "name"
	PowerUsageMilliWattsUuid        powerUsageMilliWattsTag = "uuid"
	PowerUsageMilliWattsGranularity powerUsageMilliWattsTag = "granularity"

	PowerUsageMilliWattsMinorNumber powerUsageMilliWattsField = "minor_number"
	PowerUsageMilliWattsValue       powerUsageMilliWattsField = "value"
)

type PowerUsageMilliWattsEntity struct {
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
