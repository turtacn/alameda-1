package nodeMemoryUsageBytes

import (
	metric "github.com/containers-ai/alameda/datapipe/pkg/apis/metrics/define"
	metric_dao "github.com/containers-ai/alameda/datapipe/pkg/dao/metrics"
	"github.com/containers-ai/alameda/datapipe/pkg/repositories/prometheus"
)

const (
	// MetricName Metric name to query from prometheus
	MetricName = ""
	// NodeLabel Node label name in the metric
	NodeLabel = "node"
)

// Entity Node memory usage bytes entity
type Entity struct {
	PrometheusEntity prometheus.Entity

	NodeName string
	Samples  []metric.Sample
}

// NewEntityFromPrometheusEntity New entity with field value assigned from prometheus entity
func NewEntityFromPrometheusEntity(e prometheus.Entity) Entity {

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
func (e *Entity) NodeMetric() metric_dao.NodeMetric {

	var (
		nodeMetric metric_dao.NodeMetric
	)

	nodeMetric = metric_dao.NodeMetric{
		NodeName: e.NodeName,
		Metrics: map[metric.NodeMetricType][]metric.Sample{
			metric.TypeNodeMemoryUsageBytes: e.Samples,
		},
	}

	return nodeMetric
}
