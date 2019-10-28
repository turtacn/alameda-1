package metrics

import (
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
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
	Samples       []FormatTypes.Sample
}

// NewEntityFromPrometheusEntity New entity with field value assigned from prometheus entity
func NewContainerMemoryUsageBytesEntity(e InternalPromth.Entity) ContainerMemoryUsageBytesEntity {

	var (
		samples []FormatTypes.Sample
	)

	samples = make([]FormatTypes.Sample, 0)

	for _, value := range e.Values {
		sample := FormatTypes.Sample{
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
func (e *ContainerMemoryUsageBytesEntity) ContainerMetric() DaoMetricTypes.ContainerMetric {

	var (
		containerMetric DaoMetricTypes.ContainerMetric
	)

	containerMetric = DaoMetricTypes.ContainerMetric{
		Namespace:     e.Namespace,
		PodName:       e.PodName,
		ContainerName: e.ContainerName,
		Metrics: map[FormatEnum.MetricType][]FormatTypes.Sample{
			FormatEnum.MetricTypeMemoryUsageBytes: e.Samples,
		},
	}

	return containerMetric
}
