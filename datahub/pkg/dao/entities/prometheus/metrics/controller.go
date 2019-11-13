package metrics

import (
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
)

// ControllerCPUUsageMillicoresEntity Encapsulate controller cpu usage millicores entity
type ControllerCPUUsageMillicoresEntity struct {
	NamespaceName  string
	ControllerName string
	ControllerKind string
	Samples        []FormatTypes.Sample
}

// ControllerMetric Build ControllerMetric base on entity properties
func (e *ControllerCPUUsageMillicoresEntity) ControllerMetric() DaoMetricTypes.ControllerMetric {

	m := DaoMetricTypes.ControllerMetric{
		ObjectMeta: DaoMetricTypes.ControllerObjectMeta{
			ObjectMeta: metadata.ObjectMeta{
				Namespace: e.NamespaceName,
				Name:      e.ControllerName,
			},
			Kind: e.ControllerKind,
		},
		Metrics: map[FormatEnum.MetricType][]FormatTypes.Sample{
			FormatEnum.MetricTypeCPUUsageSecondsPercentage: e.Samples,
		},
	}
	return m
}

// ControllerMemoryUsageBytesEntity Encapsulate controller memory usage millicores entity
type ControllerMemoryUsageBytesEntity struct {
	NamespaceName  string
	ControllerName string
	ControllerKind string
	Samples        []FormatTypes.Sample
}

// ControllerMetric Build ControllerMetric base on entity properties
func (e *ControllerMemoryUsageBytesEntity) ControllerMetric() DaoMetricTypes.ControllerMetric {

	m := DaoMetricTypes.ControllerMetric{
		ObjectMeta: DaoMetricTypes.ControllerObjectMeta{
			ObjectMeta: metadata.ObjectMeta{
				Namespace: e.NamespaceName,
				Name:      e.ControllerName,
			},
			Kind: e.ControllerKind,
		},
		Metrics: map[FormatEnum.MetricType][]FormatTypes.Sample{
			FormatEnum.MetricTypeMemoryUsageBytes: e.Samples,
		},
	}
	return m
}
