package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	nodeModelTimeGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "node_model_seconds",
		Help:      "Target modeling time of node",
	}, []string{"name","data_granularity"})

	nodeMAPEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "node_metric_mape",
		Help:      "MAPE of node metric",
	}, []string{"name", "metric_type", "data_granularity"})

	nodeMetricDriftCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "node_metric_drift_total",
		Help:      "Total number of node metric drift",
	}, []string{"name",  "data_granularity"})
)

type nodeMetric struct{}

func newNodeMetric() *nodeMetric {
	return &nodeMetric{}
}

func (nodeMetric *nodeMetric) setNodeMetricModelTime(
	name, dataGranularity string, val float64) {
	nodeModelTimeGauge.WithLabelValues(name, dataGranularity).Set(val)
}

func (nodeMetric *nodeMetric) setNodeMetricMAPE(
	name, metricType, dataGranularity string, val float64) {
	nodeMAPEGauge.WithLabelValues(name,
		metricType, dataGranularity).Set(val)
}

func (nodeMetric *nodeMetric) addNodeMetricDrift(
	name, dataGranularity string, val float64) {
	nodeMetricDriftCounter.WithLabelValues(name, dataGranularity).Add(val)
}
