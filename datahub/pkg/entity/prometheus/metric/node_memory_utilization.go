package metric

import (
	DaoMetric "github.com/containers-ai/alameda/datahub/pkg/dao/metric"
	"github.com/containers-ai/alameda/datahub/pkg/metric"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
)

const (
	// Metric name to query from prometheus
	NodeMemoryUtilizationMetricName = "node:node_memory_utilisation_2:"

	// Label name in prometheus metric
	NodeMemoryUtilizationLabelNode = "node"
)

// Entity Node memory usage bytes entity
type NodeMemoryUtilizationEntity struct {
	PrometheusEntity InternalPromth.Entity

	NodeName string
	Samples  []metric.Sample
}

// NewEntityFromPrometheusEntity New entity with field value assigned from prometheus entity
func NewNodeMemoryUtilizationEntity(e InternalPromth.Entity) NodeMemoryUtilizationEntity {

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

	return NodeMemoryUtilizationEntity{
		PrometheusEntity: e,
		NodeName:         e.Labels[NodeMemoryUtilizationLabelNode],
		Samples:          samples,
	}
}

// NodeMetric Build NodeMetric base on entity properties
func (e *NodeMemoryUtilizationEntity) NodeMetric() DaoMetric.NodeMetric {

	var (
		nodeMetric DaoMetric.NodeMetric
	)

	nodeMetric = DaoMetric.NodeMetric{
		NodeName: e.NodeName,
		Metrics: map[metric.NodeMetricType][]metric.Sample{
			metric.TypeNodeMemoryUsageBytes: e.Samples,
		},
	}

	return nodeMetric
}
