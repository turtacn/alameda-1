package containerMemoryUsageBytes

import (
	metric "github.com/containers-ai/alameda/datapipe/pkg/apis/metrics/define"
	DaoMetric "github.com/containers-ai/alameda/datapipe/pkg/dao/metrics"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
)

const (
	// MetricName Metric name to query from prometheus
	MetricName = "container_memory_usage_bytes"
	// NamespaceLabel Namespace label name in the metric
	NamespaceLabel = "namespace"
	// PodLabelName pod label name in the metric
	PodLabelName = "pod_name"
	// ContainerLabel container label name in the metric
	ContainerLabel = "container_name"
)

// Entity Container memory usage bytes entity
type Entity struct {
	PrometheusEntity InternalPromth.Entity

	Namespace     string
	PodName       string
	ContainerName string
	Samples       []metric.Sample
}

// NewEntityFromPrometheusEntity New entity with field value assigned from prometheus entity
func NewEntityFromPrometheusEntity(e InternalPromth.Entity) Entity {

	var (
		samples []metric.Sample
	)

	samples = make([]metric.Sample, 0)

	for _, value := range e.Values {
		sample := metric.Sample{
			Timestamp: value.UnixTime,
			Value:     value.SampleValue,
		}
		samples = append(samples, sample)
	}

	return Entity{
		PrometheusEntity: e,
		Namespace:        e.Labels[NamespaceLabel],
		PodName:          e.Labels[PodLabelName],
		ContainerName:    e.Labels[ContainerLabel],
		Samples:          samples,
	}
}

// ContainerMetric Build ContainerMetric base on entity properties
func (e *Entity) ContainerMetric() DaoMetric.ContainerMetric {

	var (
		containerMetric DaoMetric.ContainerMetric
	)

	containerMetric = DaoMetric.ContainerMetric{
		Namespace:     e.Namespace,
		PodName:       e.PodName,
		ContainerName: e.ContainerName,
		Metrics: map[metric.ContainerMetricType][]metric.Sample{
			metric.TypeContainerMemoryUsageBytes: e.Samples,
		},
	}

	return containerMetric
}
