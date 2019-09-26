package node

import (
	"strconv"
	"time"

	"github.com/containers-ai/alameda/datahub/pkg/dao/prediction"
	"github.com/containers-ai/alameda/datahub/pkg/metric"
	"github.com/containers-ai/alameda/datahub/pkg/utils"
)

type Field = string
type Tag = string
type MetricType = string

const (
	Time        Tag = "time"
	Name        Tag = "name"
	Metric      Tag = "metric"
	IsScheduled Tag = "is_scheduled"
	Granularity Tag = "granularity"
	Kind        Tag = "kind"

	ModelId      Field = "model_id"
	PredictionId Field = "prediction_id"
	Value        Field = "value"
)

var (
	// Tags Tags' name in influxdb
	Tags = []Tag{Name, Metric, IsScheduled, Granularity, Kind}
	// Fields Fields' name in influxdb
	Fields = []Field{ModelId, PredictionId, Value}
	// MetricTypeCPUUsage Enum of tag "metric"
	MetricTypeCPUUsage MetricType = "cpu_usage_seconds_percentage"
	// MetricTypeMemoryUsage Enum of tag "metric"
	MetricTypeMemoryUsage MetricType = "memory_usage_bytes"

	// LocalMetricTypeToPkgMetricType Convert local package metric type to package alameda.datahub.metric.NodeMetricType
	LocalMetricTypeToPkgMetricType = map[MetricType]metric.NodeMetricType{
		MetricTypeCPUUsage:    metric.TypeNodeCPUUsageSecondsPercentage,
		MetricTypeMemoryUsage: metric.TypeNodeMemoryUsageBytes,
	}

	// PkgMetricTypeToLocalMetricType Convert package alameda.datahub.metric.NodeMetricType to local package metric type
	PkgMetricTypeToLocalMetricType = map[metric.NodeMetricType]MetricType{
		metric.TypeNodeCPUUsageSecondsPercentage: MetricTypeCPUUsage,
		metric.TypeNodeMemoryUsageBytes:          MetricTypeMemoryUsage,
	}
)

// Entity Container prediction entity in influxDB
type Entity struct {
	Timestamp   time.Time
	Name        *string
	Metric      *MetricType
	IsScheduled *string

	ModelId      *string
	PredictionId *string
	Value        *string
}

// NewEntityFromMap Build entity from map
func NewEntityFromMap(data map[string]string) Entity {
	// TODO: log error
	tempTimestamp, _ := utils.ParseTime(data[Time])

	entity := Entity{
		Timestamp: tempTimestamp,
	}

	// InfluxDB tags
	if name, exist := data[Name]; exist {
		entity.Name = &name
	}
	if metricData, exist := data[Metric]; exist {
		entity.Metric = &metricData
	}
	if isScheduled, exist := data[IsScheduled]; exist {
		entity.IsScheduled = &isScheduled
	}

	// InfluxDB fields
	if value, exist := data[ModelId]; exist {
		entity.ModelId = &value
	}
	if value, exist := data[PredictionId]; exist {
		entity.PredictionId = &value
	}
	if value, exist := data[Value]; exist {
		entity.Value = &value
	}

	return entity
}

// NodePrediction Create container prediction base on entity
func (e Entity) NodePrediction() prediction.NodePrediction {
	var (
		isScheduled    bool
		samples        []metric.Sample
		nodePrediction prediction.NodePrediction
	)

	// TODO: log error
	samples = append(samples, metric.Sample{Timestamp: e.Timestamp, Value: *e.Value})

	nodePrediction = prediction.NodePrediction{
		NodeName:    *e.Name,
		Predictions: map[metric.NodeMetricType][]metric.Sample{},
	}

	if e.IsScheduled != nil {
		isScheduled, _ = strconv.ParseBool(*e.IsScheduled)
		nodePrediction.IsScheduled = isScheduled
	}

	metricType := LocalMetricTypeToPkgMetricType[*e.Metric]
	nodePrediction.Predictions[metricType] = samples

	return nodePrediction
}
