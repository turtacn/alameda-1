package containerCPUUsagePercentage

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/metric"
	"github.com/containers-ai/alameda/datahub/pkg/repository/prometheus"
)

const (
	// MetricName Metric name to query from prometheus
	MetricName = "namespace_pod_name_container_name:container_cpu_usage_seconds_total:sum_rate"
	// NamespaceLabel Namespace label name in the metric
	NamespaceLabel = "namespace"
	// PodLabelName pod label name in the metric
	PodLabelName = "pod_name"
	// ContainerLabel container label name in the metric
	ContainerLabel = "container_name"
)

// Entity Container cpu usage percentage entity
type Entity struct {
	PrometheusEntity prometheus.Entity

	Namespace     string
	PodName       string
	ContainerName string
	Samples       []metric.Sample
}

// NewEntityFromPrometheusEntity New entity with field value assigned from prometheus entity
func NewEntityFromPrometheusEntity(e prometheus.Entity) Entity {

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
func (e *Entity) ContainerMetric() metric.ContainerMetric {

	var (
		containerMetric metric.ContainerMetric
	)

	containerMetric = metric.ContainerMetric{
		Namespace:     e.Namespace,
		PodName:       e.PodName,
		ContainerName: e.ContainerName,
		CPUMetircs:    e.Samples,
	}

	return containerMetric
}
