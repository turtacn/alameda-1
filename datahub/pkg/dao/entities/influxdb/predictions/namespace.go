package predictions

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"time"
)

const (
	NamespaceTime        influxdb.Tag = "time"
	NamespaceName        influxdb.Tag = "name"
	NamespaceClusterName influxdb.Tag = "cluster_name"
	NamespaceMetric      influxdb.Tag = "metric"
	NamespaceMetricType  influxdb.Tag = "kind"
	NamespaceGranularity influxdb.Tag = "granularity"

	NamespaceModelId      influxdb.Field = "model_id"
	NamespacePredictionId influxdb.Field = "prediction_id"
	NamespaceValue        influxdb.Field = "value"
)

var (
	NamespaceTags = []influxdb.Tag{
		NamespaceName,
		NamespaceClusterName,
		NamespaceMetric,
		NamespaceMetricType,
		NamespaceGranularity,
	}

	NamespaceFields = []influxdb.Field{
		NamespaceModelId,
		NamespacePredictionId,
		NamespaceValue,
	}
)

// Entity Container prediction entity in influxDB
type NamespaceEntity struct {
	Time        time.Time
	Name        *string
	ClusterName *string
	Metric      *string
	MetricType  *string
	Granularity *string

	ModelId      *string
	PredictionId *string
	Value        *string
}

// NewEntityFromMap Build entity from map
func NewNamespaceEntity(data map[string]string) NamespaceEntity {
	entity := NamespaceEntity{}

	// TODO: log error
	tempTimestamp, _ := utils.ParseTime(data[string(NamespaceTime)])
	entity.Time = tempTimestamp

	// InfluxDB tags
	if value, exist := data[string(NamespaceName)]; exist {
		entity.Name = &value
	}
	if value, exist := data[string(NamespaceClusterName)]; exist {
		entity.ClusterName = &value
	}
	if value, exist := data[string(NamespaceMetric)]; exist {
		entity.Metric = &value
	}
	if value, exist := data[string(NamespaceMetricType)]; exist {
		entity.MetricType = &value
	}
	if value, exist := data[string(NamespaceGranularity)]; exist {
		entity.Granularity = &value
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
