package metrics

import (
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	K8sMetadata "github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
)

const (
	// Metric name to query from prometheus
	NodeCpuUsagePercentageMetricNameSum = "node:node_num_cpu:sum"
	NodeCpuUsagePercentageMetricNameAvg = "node:node_cpu_utilisation:avg1m"

	// Label name in prometheus metric
	NodeCpuUsagePercentageLabelNode = "node"
)

// Entity node cpu usage percentage entity
type NodeCpuUsagePercentageEntity struct {
	PrometheusEntity InternalPromth.Entity

	NodeName string
	Samples  []FormatTypes.Sample
}

// NewEntityFromPrometheusEntity New entity with field value assigned from prometheus entity
func NewNodeCpuUsagePercentageEntity(e InternalPromth.Entity) NodeCpuUsagePercentageEntity {

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

	return NodeCpuUsagePercentageEntity{
		PrometheusEntity: e,
		NodeName:         e.Labels[NodeCpuUsagePercentageLabelNode],
		Samples:          samples,
	}
}

// NodeMetric Build NodeMetric base on entity properties
func (e *NodeCpuUsagePercentageEntity) NodeMetric() DaoMetricTypes.NodeMetric {

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
