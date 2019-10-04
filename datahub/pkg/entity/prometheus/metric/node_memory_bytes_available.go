package metric

import (
	DaoMetric "github.com/containers-ai/alameda/datahub/pkg/dao/metric"
	"github.com/containers-ai/alameda/datahub/pkg/metric"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
)

const (
	// Metric name to query from prometheus
	NodeMemoryBytesAvailableMetricName = "node:node_memory_bytes_available:sum"

	// Label name in prometheus metric
	NodeMemoryBytesAvailableLabelNode = "node"
)

// Entity Node memory available entity
type NodeMemoryBytesAvailableEntity struct {
	PrometheusEntity InternalPromth.Entity

	NodeName string
	Samples  []metric.Sample
}

// NewEntityFromPrometheusEntity New entity with field value assigned from prometheus entity
func NewNodeMemoryBytesAvailableEntity(e InternalPromth.Entity) NodeMemoryBytesAvailableEntity {

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

	return NodeMemoryBytesAvailableEntity{
		PrometheusEntity: e,
		NodeName:         e.Labels[NodeMemoryBytesAvailableLabelNode],
		Samples:          samples,
	}
}

// NodeMetric Build NodeMetric base on entity properties
func (e *NodeMemoryBytesAvailableEntity) NodeMetric() DaoMetric.NodeMetric {

	var (
		nodeMetric DaoMetric.NodeMetric
	)

	nodeMetric = DaoMetric.NodeMetric{
		NodeName: e.NodeName,
		Metrics: map[metric.NodeMetricType][]metric.Sample{
			metric.TypeNodeMemoryAvailableBytes: e.Samples,
		},
	}

	return nodeMetric
}
