package nodeCPUUsagePercentage

import (
	metric "github.com/containers-ai/alameda/datapipe/pkg/apis/metrics/define"
	DaoMetric "github.com/containers-ai/alameda/datapipe/pkg/dao/metrics"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
)

const (
	// MetricName Metric name to query from prometheus
	MetricName = "node:node_cpu_utilisation:avg1m"
	// NodeLabel Node label name in the metric
	NodeLabel = "node"
)

// Entity node cpu usage percentage entity
type Entity struct {
	PrometheusEntity InternalPromth.Entity

	NodeName string
	Samples  []metric.Sample
}

// NewEntityFromPrometheusEntity New entity with field value assigned from prometheus entity
func NewEntityFromPrometheusEntity(e InternalPromth.Entity) Entity {

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

	return Entity{
		PrometheusEntity: e,
		NodeName:         e.Labels[NodeLabel],
		Samples:          samples,
	}
}

// NodeMetric Build NodeMetric base on entity properties
func (e *Entity) NodeMetric() DaoMetric.NodeMetric {

	var (
		nodeMetric DaoMetric.NodeMetric
	)

	nodeMetric = DaoMetric.NodeMetric{
		NodeName: e.NodeName,
		Metrics: map[metric.NodeMetricType][]metric.Sample{
			metric.TypeNodeCPUUsageSecondsPercentage: e.Samples,
		},
	}

	return nodeMetric
}
