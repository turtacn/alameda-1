package predictions

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"time"
)

const (
	ClusterTime        influxdb.Tag = "time"
	ClusterName        influxdb.Tag = "name"
	ClusterMetric      influxdb.Tag = "metric"
	ClusterMetricType  influxdb.Tag = "kind"
	ClusterGranularity influxdb.Tag = "granularity"

	ClusterModelId      influxdb.Field = "model_id"
	ClusterPredictionId influxdb.Field = "prediction_id"
	ClusterValue        influxdb.Field = "value"
)

var (
	ClusterTags = []influxdb.Tag{
		ClusterName,
		ClusterMetric,
		ClusterMetricType,
		ClusterGranularity,
	}

	ClusterFields = []influxdb.Field{
		ClusterModelId,
		ClusterPredictionId,
		ClusterValue,
	}
)

// Entity Container prediction entity in influxDB
type ClusterEntity struct {
	Time        time.Time
	Name        *string
	Metric      *string
	MetricType  *string
	Granularity *string

	ModelId      *string
	PredictionId *string
	Value        *string
}

// NewEntityFromMap Build entity from map
func NewClusterEntity(data map[string]string) ClusterEntity {
	entity := ClusterEntity{}

	// TODO: log error
	tempTimestamp, _ := utils.ParseTime(data[string(ClusterTime)])
	entity.Time = tempTimestamp

	// InfluxDB tags
	if value, exist := data[string(ClusterName)]; exist {
		entity.Name = &value
	}
	if value, exist := data[string(ClusterMetric)]; exist {
		entity.Metric = &value
	}
	if value, exist := data[string(ClusterMetricType)]; exist {
		entity.MetricType = &value
	}
	if value, exist := data[string(ClusterGranularity)]; exist {
		entity.Granularity = &value
	}

	// InfluxDB fields
	if value, exist := data[string(ClusterModelId)]; exist {
		entity.ModelId = &value
	}
	if value, exist := data[string(ClusterPredictionId)]; exist {
		entity.PredictionId = &value
	}
	if value, exist := data[string(ClusterValue)]; exist {
		entity.Value = &value
	}

	return entity
}
