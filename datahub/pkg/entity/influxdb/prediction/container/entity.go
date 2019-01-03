package container

import (
	"time"

	"github.com/containers-ai/alameda/datahub/pkg/dao/prediction"
)

type Field = string
type Tag = string
type MetricType = string

const (
	Database = "prediction"

	Measurement = "alameda_container_prediction"

	Time      Tag = "time"
	Namespace Tag = "namespace"
	PodName   Tag = "pod_name"
	Name      Tag = "name"
	Metric    Tag = "metric"

	Value Field = "value"
)

var (
	// Tags Tags' name in influxdb
	Tags = []Tag{Namespace, PodName, Name, Metric}
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

	Namespace string
	PodName   string
	Name      string
	Metric    MetricType
	Value     string
}

// ContainerPrediction Create container prediction base on entity
func (e Entity) ContainerPrediction() prediction.ContainerPrediction {

	var (
		samples             []prediction.Sample
		containerPrediction prediction.ContainerPrediction
	)

	samples = append(samples, prediction.Sample{Timestamp: e.Timestamp, Value: e.Value})

	containerPrediction = prediction.ContainerPrediction{
		Namespace:     e.Namespace,
		PodName:       e.PodName,
		ContainerName: e.Name,
	}

	switch e.Metric {
	case MetricTypeCPUUsage:
		containerPrediction.CPUPredictions = samples
	case MetricTypeMemoryUsage:
		containerPrediction.MemoryPredictions = samples
	}

	return containerPrediction
}
