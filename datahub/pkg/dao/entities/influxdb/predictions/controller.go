package predictions

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"time"
)

const (
	ControllerTime        influxdb.Tag = "time"
	ControllerName        influxdb.Tag = "name"
	ControllerNamespace   influxdb.Tag = "namespace"
	ControllerClusterName influxdb.Tag = "cluster_name"
	ControllerMetric      influxdb.Tag = "metric"
	ControllerMetricType  influxdb.Tag = "kind"
	ControllerGranularity influxdb.Tag = "granularity"
	ControllerKind        influxdb.Tag = "controller_kind"

	ControllerModelId      influxdb.Field = "model_id"
	ControllerPredictionId influxdb.Field = "prediction_id"
	ControllerValue        influxdb.Field = "value"
)

var (
	ControllerTags = []influxdb.Tag{
		ControllerName,
		ControllerNamespace,
		ControllerClusterName,
		ControllerMetric,
		ControllerMetricType,
		ControllerGranularity,
		ControllerKind,
	}

	ControllerFields = []influxdb.Field{
		ControllerModelId,
		ControllerPredictionId,
		ControllerValue,
	}
)

// Entity Container prediction entity in influxDB
type ControllerEntity struct {
	Time        time.Time
	Name        *string
	Namespace   *string
	ClusterName *string
	Metric      *string
	MetricType  *string
	Granularity *string
	Kind        *string

	ModelId      *string
	PredictionId *string
	Value        *string
}

// NewEntityFromMap Build entity from map
func NewControllerEntity(data map[string]string) ControllerEntity {
	entity := ControllerEntity{}

	// TODO: log error
	tempTimestamp, _ := utils.ParseTime(data[string(ControllerTime)])
	entity.Time = tempTimestamp

	// InfluxDB tags
	if value, exist := data[string(ControllerName)]; exist {
		entity.Name = &value
	}
	if value, exist := data[string(ControllerNamespace)]; exist {
		entity.Namespace = &value
	}
	if value, exist := data[string(ControllerClusterName)]; exist {
		entity.ClusterName = &value
	}
	if value, exist := data[string(ControllerMetric)]; exist {
		entity.Metric = &value
	}
	if value, exist := data[string(ControllerMetricType)]; exist {
		entity.MetricType = &value
	}
	if value, exist := data[string(ControllerGranularity)]; exist {
		entity.Granularity = &value
	}
	if value, exist := data[string(ControllerKind)]; exist {
		entity.Kind = &value
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
