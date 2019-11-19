package predictions

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"time"
)

const (
	ApplicationTime        influxdb.Tag = "time"
	ApplicationName        influxdb.Tag = "name"
	ApplicationNameSpace   influxdb.Tag = "namespace"
	ApplicationClusterName influxdb.Tag = "cluster_name"
	ApplicationMetric      influxdb.Tag = "metric"
	ApplicationMetricType  influxdb.Tag = "kind"
	ApplicationGranularity influxdb.Tag = "granularity"

	ApplicationModelId      influxdb.Field = "model_id"
	ApplicationPredictionId influxdb.Field = "prediction_id"
	ApplicationValue        influxdb.Field = "value"
)

var (
	ApplicationTags = []influxdb.Tag{
		ApplicationName,
		ApplicationNameSpace,
		ApplicationClusterName,
		ApplicationMetric,
		ApplicationMetricType,
		ApplicationGranularity,
	}

	ApplicationFields = []influxdb.Field{
		ApplicationModelId,
		ApplicationPredictionId,
		ApplicationValue,
	}
)

// Entity Container prediction entity in influxDB
type ApplicationEntity struct {
	Time        time.Time
	Name        *string
	Namespace   *string
	ClusterName *string
	Metric      *string
	MetricType  *string
	Granularity *string

	ModelId      *string
	PredictionId *string
	Value        *string
}

// NewEntityFromMap Build entity from map
func NewApplicationEntity(data map[string]string) ApplicationEntity {
	entity := ApplicationEntity{}

	// TODO: log error
	tempTimestamp, _ := utils.ParseTime(data[string(ApplicationTime)])
	entity.Time = tempTimestamp

	// InfluxDB tags
	if value, exist := data[string(ApplicationName)]; exist {
		entity.Name = &value
	}
	if value, exist := data[string(ApplicationNameSpace)]; exist {
		entity.Namespace = &value
	}
	if value, exist := data[string(ApplicationClusterName)]; exist {
		entity.ClusterName = &value
	}
	if value, exist := data[string(ApplicationMetric)]; exist {
		entity.Metric = &value
	}
	if value, exist := data[string(ApplicationMetricType)]; exist {
		entity.MetricType = &value
	}
	if value, exist := data[string(ApplicationGranularity)]; exist {
		entity.Granularity = &value
	}

	// InfluxDB fields
	if value, exist := data[string(ApplicationModelId)]; exist {
		entity.ModelId = &value
	}
	if value, exist := data[string(ApplicationPredictionId)]; exist {
		entity.PredictionId = &value
	}
	if value, exist := data[string(ApplicationValue)]; exist {
		entity.Value = &value
	}

	return entity
}
