package prediction

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"strconv"
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

	PowerUsageMilliWattsModelId      powerUsageMilliWattsField = "model_id"
	PowerUsageMilliWattsPredictionId powerUsageMilliWattsField = "prediction_id"
	PowerUsageMilliWattsMinorNumber  powerUsageMilliWattsField = "minor_number"
	PowerUsageMilliWattsValue        powerUsageMilliWattsField = "value"
)

type PowerUsageMilliWattsEntity struct {
	Time        time.Time
	Host        *string
	Instance    *string
	Job         *string
	Name        *string
	Uuid        *string
	Granularity *string

	ModelId      *string
	PredictionId *string
	MinorNumber  *string
	Value        *float64
}

func NewPowerUsageMilliWattsEntityFromMap(data map[string]string) PowerUsageMilliWattsEntity {
	tempTimestamp, _ := utils.ParseTime(data[PowerUsageMilliWattsTime])

	entity := PowerUsageMilliWattsEntity{
		Time: tempTimestamp,
	}

	// InfluxDB tags
	if valueStr, exist := data[PowerUsageMilliWattsHost]; exist {
		entity.Host = &valueStr
	}
	if valueStr, exist := data[PowerUsageMilliWattsInstance]; exist {
		entity.Instance = &valueStr
	}
	if valueStr, exist := data[PowerUsageMilliWattsJob]; exist {
		entity.Job = &valueStr
	}
	if valueStr, exist := data[PowerUsageMilliWattsName]; exist {
		entity.Name = &valueStr
	}
	if valueStr, exist := data[PowerUsageMilliWattsUuid]; exist {
		entity.Uuid = &valueStr
	}
	if valueStr, exist := data[PowerUsageMilliWattsGranularity]; exist {
		entity.Granularity = &valueStr
	}

	// InfluxDB fields
	if valueStr, exist := data[PowerUsageMilliWattsModelId]; exist {
		entity.ModelId = &valueStr
	}
	if valueStr, exist := data[PowerUsageMilliWattsPredictionId]; exist {
		entity.PredictionId = &valueStr
	}
	if valueStr, exist := data[PowerUsageMilliWattsMinorNumber]; exist {
		entity.MinorNumber = &valueStr
	}
	if valueFloat, exist := data[PowerUsageMilliWattsValue]; exist {
		value, _ := strconv.ParseFloat(valueFloat, 64)
		entity.Value = &value
	}

	return entity
}
