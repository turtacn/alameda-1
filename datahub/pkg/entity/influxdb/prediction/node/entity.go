package node

import (
	"strconv"
	"time"

	"github.com/containers-ai/alameda/datahub/pkg/dao/prediction"
	"github.com/containers-ai/alameda/datahub/pkg/metric"
)

type Field = string
type Tag = string
type MetricType = string

const (
	Database = "prediction"

	Measurement = "alameda_node_prediction"

	Time        Tag = "time"
	Name        Tag = "name"
	Metric      Tag = "metric"
	IsScheduled Tag = "is_scheduled"

	Value Field = "value"
)

var (
	// Tags Tags' name in influxdb
	Tags = []Tag{Name, Metric, IsScheduled}
	// Fields Fields' name in influxdb
	Fields = []Field{Value}
	// MetricTypeCPUUsage Enum of tag "metric"
	MetricTypeCPUUsage MetricType = "cpu_usage_seconds_percentage"
	// MetricTypeMemoryUsage Enum of tag "metric"
	MetricTypeMemoryUsage MetricType = "memory_usage_bytes"
)

// Entity Container prediction entity in influxDB
type Entity struct {
	Timestamp time.Time

	Name        string
	Metric      MetricType
	Value       string
	IsScheduled string
}

// NodePrediction Create container prediction base on entity
func (e Entity) NodePrediction() prediction.NodePrediction {

	var (
		isScheduled    bool
		samples        []metric.Sample
		NodePrediction prediction.NodePrediction
	)

	// TODO: log error
	isScheduled, _ = strconv.ParseBool(e.IsScheduled)
	samples = append(samples, metric.Sample{Timestamp: e.Timestamp, Value: e.Value})

	NodePrediction = prediction.NodePrediction{
		NodeName:    e.Name,
		IsScheduled: isScheduled,
	}

	switch e.Metric {
	case MetricTypeCPUUsage:
		NodePrediction.CPUUsagePredictions = samples
	case MetricTypeMemoryUsage:
		NodePrediction.MemoryUsagePredictions = samples
	}

	return NodePrediction
}
