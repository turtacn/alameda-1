package predictions

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"time"
)

const (
	NamespaceTime        influxdb.Tag = "time"
	NamespaceName        influxdb.Tag = "name"
	NamespaceMetric      influxdb.Tag = "metric"
	NamespaceGranularity influxdb.Tag = "granularity"
	NamespaceKind        influxdb.Tag = "kind"

	NamespaceModelId      influxdb.Field = "model_id"
	NamespacePredictionId influxdb.Field = "prediction_id"
	NamespaceValue        influxdb.Field = "value"
)

var (
	NamespaceTags   = []influxdb.Tag{NamespaceName, NamespaceMetric, NamespaceGranularity}
	NamespaceFields = []influxdb.Field{NamespaceModelId, NamespacePredictionId, NamespaceValue}
)

// Entity Container prediction entity in influxDB
type NamespaceEntity struct {
	Time        time.Time
	Name        *string
	Metric      *string
	Granularity *string
	Kind        *string

	ModelId      *string
	PredictionId *string
	Value        *string
}

// NewEntityFromMap Build entity from map
func NewNamespaceEntityFromMap(data map[string]string) NamespaceEntity {
	// TODO: log error
	tempTimestamp, _ := utils.ParseTime(data[string(NamespaceTime)])

	entity := NamespaceEntity{
		Time: tempTimestamp,
	}

	// InfluxDB tags
	if name, exist := data[string(NamespaceName)]; exist {
		entity.Name = &name
	}
	if metricData, exist := data[string(NamespaceMetric)]; exist {
		entity.Metric = &metricData
	}
	if valueStr, exist := data[string(NamespaceGranularity)]; exist {
		entity.Granularity = &valueStr
	}
	if valueStr, exist := data[string(NamespaceKind)]; exist {
		entity.Kind = &valueStr
	}

	// InfluxDB fields
	if value, exist := data[string(NamespaceModelId)]; exist {
		entity.ModelId = &value
	}
	if value, exist := data[string(NamespacePredictionId)]; exist {
		entity.PredictionId = &value
	}
	if value, exist := data[string(NamespaceValue)]; exist {
		entity.Value = &value
	}

	return entity
}
