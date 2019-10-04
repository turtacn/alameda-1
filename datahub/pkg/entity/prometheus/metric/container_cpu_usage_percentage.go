package metric

import (
	"fmt"
	DaoMetric "github.com/containers-ai/alameda/datahub/pkg/dao/metric"
	"github.com/containers-ai/alameda/datahub/pkg/metric"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/containers-ai/alameda/pkg/utils/log"
	"strconv"
)

const (
	// Metric name to query from prometheus
	ContainerCpuUsagePercentageMetricName = "namespace_pod_name_container_name:container_cpu_usage_seconds_total:sum_rate"

	// Label name in prometheus metric
	ContainerCpuUsagePercentageLabelNamespace     = "namespace"
	ContainerCpuUsagePercentageLabelPodName       = "pod_name"
	ContainerCpuUsagePercentageLabelContainerName = "container_name"
)

var (
	scope = log.RegisterScope("prometheus", "metrics repository", 0)
)

// Entity Container cpu usage percentage entity
type ContainerCpuUsagePercentageEntity struct {
	PrometheusEntity InternalPromth.Entity

	Namespace     string
	PodName       string
	ContainerName string
	Samples       []metric.Sample
}

// NewEntityFromPrometheusEntity New entity with field value assigned from prometheus entity
func NewContainerCpuUsagePercentageEntity(e InternalPromth.Entity) ContainerCpuUsagePercentageEntity {

	var (
		samples []metric.Sample
	)

	samples = make([]metric.Sample, 0)

	for _, value := range e.Values {
		v := "0"
		if s, err := strconv.ParseFloat(value.SampleValue, 64); err == nil {
			v = fmt.Sprintf("%f", s*1000)
		} else {
			scope.Errorf("container_cpu_usage_percentage.NewContainerCpuUsagePercentageEntity: %s", err.Error())
		}
		sample := metric.Sample{
			Timestamp: value.UnixTime,
			Value:     v,
		}
		samples = append(samples, sample)
	}

	return ContainerCpuUsagePercentageEntity{
		PrometheusEntity: e,
		Namespace:        e.Labels[ContainerCpuUsagePercentageLabelNamespace],
		PodName:          e.Labels[ContainerCpuUsagePercentageLabelPodName],
		ContainerName:    e.Labels[ContainerCpuUsagePercentageLabelContainerName],
		Samples:          samples,
	}
}

// ContainerMetric Build ContainerMetric base on entity properties
func (e *ContainerCpuUsagePercentageEntity) ContainerMetric() DaoMetric.ContainerMetric {

	var (
		containerMetric DaoMetric.ContainerMetric
	)

	containerMetric = DaoMetric.ContainerMetric{
		Namespace:     e.Namespace,
		PodName:       e.PodName,
		ContainerName: e.ContainerName,
		Metrics: map[metric.ContainerMetricType][]metric.Sample{
			metric.TypeContainerCPUUsageSecondsPercentage: e.Samples,
		},
	}

	return containerMetric
}
