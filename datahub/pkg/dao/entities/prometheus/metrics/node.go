package metrics

import (
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	K8sMetadata "github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
)

type NodeMemoryBytesUsageEntity struct {
	NodeName string
	Samples  []FormatTypes.Sample
}

// NodeMetric Build NodeMetric base on entity properties
func (e *NodeMemoryBytesUsageEntity) NodeMetric() DaoMetricTypes.NodeMetric {

	var (
		nodeMetric DaoMetricTypes.NodeMetric
	)

	nodeMetric = DaoMetricTypes.NodeMetric{
		ObjectMeta: K8sMetadata.ObjectMeta{
			Name: e.NodeName,
		},
		Metrics: map[FormatEnum.MetricType][]FormatTypes.Sample{
			FormatEnum.MetricTypeMemoryUsageBytes: e.Samples,
		},
	}

	return nodeMetric
}

type NodeCPUUsageMillicoresEntity struct {
	NodeName string
	Samples  []FormatTypes.Sample
}

// NodeMetric Build NodeMetric base on entity properties
func (e *NodeCPUUsageMillicoresEntity) NodeMetric() DaoMetricTypes.NodeMetric {

	var (
		nodeMetric DaoMetricTypes.NodeMetric
	)

	nodeMetric = DaoMetricTypes.NodeMetric{
		ObjectMeta: K8sMetadata.ObjectMeta{
			Name: e.NodeName,
		},
		Metrics: map[FormatEnum.MetricType][]FormatTypes.Sample{
			FormatEnum.MetricTypeCPUUsageSecondsPercentage: e.Samples,
		},
	}

	return nodeMetric
}
