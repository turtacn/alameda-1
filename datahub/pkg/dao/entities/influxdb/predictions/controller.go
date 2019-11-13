package predictions

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"time"
)

const (
	ControllerTime        influxdb.Tag = "time"
	ControllerName        influxdb.Tag = "name"
	ControllerMetric      influxdb.Tag = "metric"
	ControllerGranularity influxdb.Tag = "granularity"
	ControllerKind        influxdb.Tag = "kind"
	ControllerCtlKind     influxdb.Tag = "controller_kind"

	ControllerModelId      influxdb.Field = "model_id"
	ControllerPredictionId influxdb.Field = "prediction_id"
	ControllerValue        influxdb.Field = "value"
)

var (
	ControllerTags   = []influxdb.Tag{ControllerName, ControllerMetric, ControllerGranularity, ControllerKind}
	ControllerFields = []influxdb.Field{ControllerModelId, ControllerPredictionId, ControllerValue}
)

// Entity Container prediction entity in influxDB
type ControllerEntity struct {
	Time        time.Time
	Name        *string
	Metric      *string
	Kind        *string
	Granularity *string
	CtlKind     *string

	ModelId      *string
	PredictionId *string
	Value        *string
}

// NewEntityFromMap Build entity from map
func NewControllerEntityFromMap(data map[string]string) ControllerEntity {
	// TODO: log error
	tempTimestamp, _ := utils.ParseTime(data[string(ControllerTime)])

	entity := ControllerEntity{
		Time: tempTimestamp,
	}

	// InfluxDB tags
	if name, exist := data[string(ControllerName)]; exist {
		entity.Name = &name
	}
	if metricData, exist := data[string(ControllerMetric)]; exist {
		entity.Metric = &metricData
	}
	if valueStr, exist := data[string(ControllerGranularity)]; exist {
		entity.Granularity = &valueStr
	}
	if valueStr, exist := data[string(ControllerKind)]; exist {
		entity.Kind = &valueStr
	}
	if valueStr, exist := data[string(ControllerCtlKind)]; exist {
		entity.CtlKind = &valueStr
	}

	// InfluxDB fields
	if value, exist := data[string(ControllerModelId)]; exist {
		entity.ModelId = &value
	}
	if value, exist := data[string(ControllerPredictionId)]; exist {
		entity.PredictionId = &value
	}
	if value, exist := data[string(ControllerValue)]; exist {
		entity.Value = &value
	}

	return entity
}
