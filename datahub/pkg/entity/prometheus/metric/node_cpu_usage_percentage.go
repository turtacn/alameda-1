package metric

import (
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/metric/types"
	"github.com/containers-ai/alameda/datahub/pkg/metric"
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
	Samples  []metric.Sample
}

// NewEntityFromPrometheusEntity New entity with field value assigned from prometheus entity
func NewNodeCpuUsagePercentageEntity(e InternalPromth.Entity) NodeCpuUsagePercentageEntity {

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
		NodeName: e.NodeName,
		Metrics: map[metric.NodeMetricType][]metric.Sample{
			metric.TypeNodeCPUUsageSecondsPercentage: e.Samples,
		},
	}

	return nodeMetric
}
