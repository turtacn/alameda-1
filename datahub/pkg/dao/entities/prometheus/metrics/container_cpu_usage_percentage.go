package metrics

import (
	"fmt"
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
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
	scope = Log.RegisterScope("prometheus", "metrics repository", 0)
)

// Entity Container cpu usage percentage entity
type ContainerCpuUsagePercentageEntity struct {
	PrometheusEntity InternalPromth.Entity

	Namespace     string
	PodName       string
	ContainerName string
	Samples       []FormatTypes.Sample
}

// NewEntityFromPrometheusEntity New entity with field value assigned from prometheus entity
func NewContainerCpuUsagePercentageEntity(e InternalPromth.Entity) ContainerCpuUsagePercentageEntity {

	var (
		samples []FormatTypes.Sample
	)

	samples = make([]FormatTypes.Sample, 0)

	for _, value := range e.Values {
		v := "0"
		if s, err := strconv.ParseFloat(value.SampleValue, 64); err == nil {
			v = fmt.Sprintf("%f", s*1000)
		} else {
			scope.Errorf("container_cpu_usage_percentage.NewContainerCpuUsagePercentageEntity: %s", err.Error())
		}
		sample := FormatTypes.Sample{
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
func (e *ContainerCpuUsagePercentageEntity) ContainerMetric() DaoMetricTypes.ContainerMetric {

	var (
		containerMetric DaoMetricTypes.ContainerMetric
	)

	containerMetric = DaoMetricTypes.ContainerMetric{
		Namespace:     e.Namespace,
		PodName:       e.PodName,
		ContainerName: e.ContainerName,
		Metrics: map[FormatEnum.MetricType][]FormatTypes.Sample{
			FormatEnum.MetricTypeCPUUsageSecondsPercentage: e.Samples,
		},
	}

	return containerMetric
}
