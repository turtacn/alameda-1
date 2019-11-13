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
	ClusterGranularity influxdb.Tag = "granularity"
	ClusterKind        influxdb.Tag = "kind"

	ClusterModelId      influxdb.Field = "model_id"
	ClusterPredictionId influxdb.Field = "prediction_id"
	ClusterValue        influxdb.Field = "value"
)

var (
	ClusterTags   = []influxdb.Tag{ClusterName, ClusterMetric, ClusterGranularity}
	ClusterFields = []influxdb.Field{ClusterModelId, ClusterPredictionId, ClusterValue}
)

// Entity Container prediction entity in influxDB
type ClusterEntity struct {
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
func NewClusterEntityFromMap(data map[string]string) ClusterEntity {
	// TODO: log error
	tempTimestamp, _ := utils.ParseTime(data[string(ClusterTime)])

	entity := ClusterEntity{
		Time: tempTimestamp,
	}

	// InfluxDB tags
	if name, exist := data[string(ClusterName)]; exist {
		entity.Name = &name
	}
	if metricData, exist := data[string(ClusterMetric)]; exist {
		entity.Metric = &metricData
	}
	if valueStr, exist := data[string(ClusterGranularity)]; exist {
		entity.Granularity = &valueStr
	}
	if valueStr, exist := data[string(ClusterKind)]; exist {
		entity.Kind = &valueStr
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
