package metrics

import (
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
)

// AppCPUUsageMillicoresEntity Encapsulate App cpu usage millicores entity
type AppCPUUsageMillicoresEntity struct {
	NamespaceName string
	AppName       string
	Samples       []FormatTypes.Sample
}

// AppMetric Build AppMetric base on entity properties
func (e *AppCPUUsageMillicoresEntity) AppMetric() DaoMetricTypes.AppMetric {

	m := DaoMetricTypes.AppMetric{
		ObjectMeta: metadata.ObjectMeta{
			Namespace: e.NamespaceName,
			Name:      e.AppName,
		},
		Metrics: map[FormatEnum.MetricType][]FormatTypes.Sample{
			FormatEnum.MetricTypeCPUUsageSecondsPercentage: e.Samples,
		},
	}
	return m
}

// AppMemoryUsageBytesEntity Encapsulate App memory usage millicores entity
type AppMemoryUsageBytesEntity struct {
	NamespaceName string
	AppName       string
	Samples       []FormatTypes.Sample
}

// AppMetric Build AppMetric base on entity properties
func (e *AppMemoryUsageBytesEntity) AppMetric() DaoMetricTypes.AppMetric {

	m := DaoMetricTypes.AppMetric{
		ObjectMeta: metadata.ObjectMeta{
			Namespace: e.NamespaceName,
			Name:      e.AppName,
		},
		Metrics: map[FormatEnum.MetricType][]FormatTypes.Sample{
			FormatEnum.MetricTypeMemoryUsageBytes: e.Samples,
		},
	}
	return m
}
