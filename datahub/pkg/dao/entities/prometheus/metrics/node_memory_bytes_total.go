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
	NodeMemoryBytesTotalMetricName = "node:node_memory_bytes_total:sum"

	// Label name in prometheus metric
	NodeMemoryBytesTotalLabelNode = "node"
)

// Entity Node total memory entity
type NodeMemoryBytesTotalEntity struct {
	PrometheusEntity InternalPromth.Entity

	NodeName string
	Samples  []FormatTypes.Sample
}

// NewEntityFromPrometheusEntity New entity with field value assigned from prometheus entity
func NewNodeMemoryBytesTotalEntity(e InternalPromth.Entity) NodeMemoryBytesTotalEntity {

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

	return NodeMemoryBytesTotalEntity{
		PrometheusEntity: e,
		NodeName:         e.Labels[NodeMemoryBytesTotalLabelNode],
		Samples:          samples,
	}
}

// NodeMetric Build NodeMetric base on entity properties
func (e *NodeMemoryBytesTotalEntity) NodeMetric() DaoMetricTypes.NodeMetric {

	var (
		nodeMetric DaoMetricTypes.NodeMetric
	)

	nodeMetric = DaoMetricTypes.NodeMetric{
		ObjectMeta: K8sMetadata.ObjectMeta{
			Name: e.NodeName,
		},
		Metrics: map[FormatEnum.MetricType][]FormatTypes.Sample{
			FormatEnum.MetricTypeMemoryTotalBytes: e.Samples,
		},
	}

	return nodeMetric
}
