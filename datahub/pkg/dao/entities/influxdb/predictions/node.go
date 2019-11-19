package predictions

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"time"
)

const (
	NodeTime        influxdb.Tag = "time"
	NodeName        influxdb.Tag = "name"
	NodeClusterName influxdb.Tag = "cluster_name"
	NodeMetric      influxdb.Tag = "metric"
	NodeMetricType  influxdb.Tag = "kind"
	NodeGranularity influxdb.Tag = "granularity"
	NodeIsScheduled influxdb.Tag = "is_scheduled"

	NodeModelId      influxdb.Field = "model_id"
	NodePredictionId influxdb.Field = "prediction_id"
	NodeValue        influxdb.Field = "value"
)

var (
	NodeTags = []influxdb.Tag{
		NodeName,
		NodeClusterName,
		NodeMetric,
		NodeMetricType,
		NodeGranularity,
		NodeIsScheduled,
	}

	NodeFields = []influxdb.Field{
		NodeModelId,
		NodePredictionId,
		NodeValue,
	}
)

// Entity Container prediction entity in influxDB
type NodeEntity struct {
	Time        time.Time
	Name        *string
	ClusterName *string
	Metric      *string
	MetricType  *string
	Granularity *string
	IsScheduled *string

	ModelId      *string
	PredictionId *string
	Value        *string
}

// NewEntityFromMap Build entity from map
func NewNodeEntity(data map[string]string) NodeEntity {
	entity := NodeEntity{}

	// TODO: log error
	tempTimestamp, _ := utils.ParseTime(data[string(NodeTime)])
	entity.Time = tempTimestamp

	// InfluxDB tags
	if value, exist := data[string(NodeName)]; exist {
		entity.Name = &value
	}
	if value, exist := data[string(NodeClusterName)]; exist {
		entity.ClusterName = &value
	}
	if value, exist := data[string(NodeMetric)]; exist {
		entity.Metric = &value
	}
	if value, exist := data[string(NodeMetricType)]; exist {
		entity.MetricType = &value
	}
	if value, exist := data[string(NodeGranularity)]; exist {
		entity.Granularity = &value
	}
	if value, exist := data[string(NodeIsScheduled)]; exist {
		entity.IsScheduled = &value
	}

	// InfluxDB fields
	if value, exist := data[string(NodeModelId)]; exist {
		entity.ModelId = &value
	}
	if value, exist := data[string(NodePredictionId)]; exist {
		entity.PredictionId = &value
	}
	if value, exist := data[string(NodeValue)]; exist {
		entity.Value = &value
	}

	return entity
}
