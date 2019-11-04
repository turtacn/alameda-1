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
	NodeMemoryBytesAvailableMetricName = "node:node_memory_bytes_available:sum"

	// Label name in prometheus metric
	NodeMemoryBytesAvailableLabelNode = "node"
)

// Entity Node memory available entity
type NodeMemoryBytesAvailableEntity struct {
	PrometheusEntity InternalPromth.Entity

	NodeName string
	Samples  []FormatTypes.Sample
}

// NewEntityFromPrometheusEntity New entity with field value assigned from prometheus entity
func NewNodeMemoryBytesAvailableEntity(e InternalPromth.Entity) NodeMemoryBytesAvailableEntity {

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

	return NodeMemoryBytesAvailableEntity{
		PrometheusEntity: e,
		NodeName:         e.Labels[NodeMemoryBytesAvailableLabelNode],
		Samples:          samples,
	}
}

// NodeMetric Build NodeMetric base on entity properties
func (e *NodeMemoryBytesAvailableEntity) NodeMetric() DaoMetricTypes.NodeMetric {

	var (
		nodeMetric DaoMetricTypes.NodeMetric
	)

	nodeMetric = DaoMetricTypes.NodeMetric{
		ObjectMeta: K8sMetadata.ObjectMeta{
			Name: e.NodeName,
		},
		Metrics: map[FormatEnum.MetricType][]FormatTypes.Sample{
			FormatEnum.MetricTypeMemoryAvailableBytes: e.Samples,
		},
	}

	return nodeMetric
}
