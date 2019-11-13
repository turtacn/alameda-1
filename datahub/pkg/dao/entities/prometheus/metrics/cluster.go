package metrics

import (
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
)

type ClusterCPUUsageMillicoresEntity struct {
	ClusterName string
	Samples     []FormatTypes.Sample
}

// ClusterMetric Build ClusterMetric base on entity properties
func (e *ClusterCPUUsageMillicoresEntity) ClusterMetric() DaoMetricTypes.ClusterMetric {

	m := DaoMetricTypes.ClusterMetric{
		ObjectMeta: metadata.ObjectMeta{
			Name: e.ClusterName,
		},
		Metrics: map[FormatEnum.MetricType][]FormatTypes.Sample{
			FormatEnum.MetricTypeCPUUsageSecondsPercentage: e.Samples,
		},
	}
	return m
}

type ClusterMemoryUsageBytesEntity struct {
	ClusterName string
	Samples     []FormatTypes.Sample
}

// ClusterMetric Build ClusterMetric base on entity properties
func (e *ClusterMemoryUsageBytesEntity) ClusterMetric() DaoMetricTypes.ClusterMetric {

	m := DaoMetricTypes.ClusterMetric{
		ObjectMeta: metadata.ObjectMeta{
			Name: e.ClusterName,
		},
		Metrics: map[FormatEnum.MetricType][]FormatTypes.Sample{
			FormatEnum.MetricTypeMemoryUsageBytes: e.Samples,
		},
	}
	return m
}
