package metric

import (
	DaoMetric "github.com/containers-ai/alameda/datahub/pkg/dao/metric"
	"github.com/containers-ai/alameda/datahub/pkg/metric"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
)

const (
	// Metric name to query from prometheus
	ContainerMemoryUsageBytesMetricName = "container_memory_usage_bytes"

	// Label name in prometheus metric
	ContainerMemoryUsageBytesLabelNamespace     = "namespace"
	ContainerMemoryUsageBytesLabelPodName       = "pod_name"
	ContainerMemoryUsageBytesLabelContainerName = "container_name"
)

// Entity Container memory usage bytes entity
type ContainerMemoryUsageBytesEntity struct {
	PrometheusEntity InternalPromth.Entity

	Namespace     string
	PodName       string
	ContainerName string
	Samples       []metric.Sample
}

// NewEntityFromPrometheusEntity New entity with field value assigned from prometheus entity
func NewContainerMemoryUsageBytesEntity(e InternalPromth.Entity) ContainerMemoryUsageBytesEntity {

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

	return ContainerMemoryUsageBytesEntity{
		PrometheusEntity: e,
		Namespace:        e.Labels[ContainerMemoryUsageBytesLabelNamespace],
		PodName:          e.Labels[ContainerMemoryUsageBytesLabelPodName],
		ContainerName:    e.Labels[ContainerMemoryUsageBytesLabelContainerName],
		Samples:          samples,
	}
}

// ContainerMetric Build ContainerMetric base on entity properties
func (e *ContainerMemoryUsageBytesEntity) ContainerMetric() DaoMetric.ContainerMetric {

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
