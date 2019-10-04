package metric

import (
	DaoMetric "github.com/containers-ai/alameda/datahub/pkg/dao/metric"
	"github.com/containers-ai/alameda/datahub/pkg/metric"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
)

const (
	// Metric name to query from prometheus
	NodeMemoryBytesTotalMetricName = "node:node_memory_bytes_total:sum"

	// Label name in prometheus metric
	NodeMemoryBytesTotalLabelNode = "node"
)

// Entity Node total memory entity
type NodeMemoryBytesTotalEntity struct {
	PrometheusEntity InternalPromth.Entity

	NodeName string
	Samples  []metric.Sample
}

// NewEntityFromPrometheusEntity New entity with field value assigned from prometheus entity
func NewNodeMemoryBytesTotalEntity(e InternalPromth.Entity) NodeMemoryBytesTotalEntity {

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

	return NodeMemoryBytesTotalEntity{
		PrometheusEntity: e,
		NodeName:         e.Labels[NodeMemoryBytesTotalLabelNode],
		Samples:          samples,
	}
}

// NodeMetric Build NodeMetric base on entity properties
func (e *NodeMemoryBytesTotalEntity) NodeMetric() DaoMetric.NodeMetric {

	var (
		nodeMetric DaoMetric.NodeMetric
	)

	nodeMetric = DaoMetric.NodeMetric{
		NodeName: e.NodeName,
		Metrics: map[metric.NodeMetricType][]metric.Sample{
			metric.TypeNodeMemoryTotalBytes: e.Samples,
		},
	}

	return nodeMetric
}
