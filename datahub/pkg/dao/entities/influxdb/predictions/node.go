package predictions

import (
	"github.com/containers-ai/alameda/datahub/pkg/utils"
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"time"
)

const (
	NodeTime        influxdb.Tag = "time"
	NodeName        influxdb.Tag = "name"
	NodeMetric      influxdb.Tag = "metric"
	NodeIsScheduled influxdb.Tag = "is_scheduled"
	NodeGranularity influxdb.Tag = "granularity"
	NodeKind        influxdb.Tag = "kind"

	NodeModelId      influxdb.Field = "model_id"
	NodePredictionId influxdb.Field = "prediction_id"
	NodeValue        influxdb.Field = "value"
)

var (
	NodeTags   = []influxdb.Tag{NodeName, NodeMetric, NodeIsScheduled, NodeGranularity, NodeKind}
	NodeFields = []influxdb.Field{NodeModelId, NodePredictionId, NodeValue}
)

// Entity Container prediction entity in influxDB
type NodeEntity struct {
	Time        time.Time
	Name        *string
	Metric      *string
	IsScheduled *string
	Granularity *string
	Kind        *string

	ModelId      *string
	PredictionId *string
	Value        *string
}

// NewEntityFromMap Build entity from map
func NewNodeEntityFromMap(data map[string]string) NodeEntity {
	// TODO: log error
	tempTimestamp, _ := utils.ParseTime(data[string(NodeTime)])

	entity := NodeEntity{
		Time: tempTimestamp,
	}

	// InfluxDB tags
	if name, exist := data[string(NodeName)]; exist {
		entity.Name = &name
	}
	if metricData, exist := data[string(NodeMetric)]; exist {
		entity.Metric = &metricData
	}
	if isScheduled, exist := data[string(NodeIsScheduled)]; exist {
		entity.IsScheduled = &isScheduled
	}
	if valueStr, exist := data[string(NodeGranularity)]; exist {
		entity.Granularity = &valueStr
	}
	if valueStr, exist := data[string(NodeKind)]; exist {
		entity.Kind = &valueStr
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
