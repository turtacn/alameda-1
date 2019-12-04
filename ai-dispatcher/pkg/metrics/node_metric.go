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
	}, []string{"name", "data_granularity", "export_timestamp"})

	nodeModelTimeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "node_model_seconds_total",
		Help:      "Total target modeling time of node",
	}, []string{"name", "data_granularity", "export_timestamp"})

	nodeMAPEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "node_metric_mape",
		Help:      "MAPE of node metric",
	}, []string{"name", "metric_type", "data_granularity", "export_timestamp"})

	nodeRMSEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "node_metric_rmse",
		Help:      "RMSE of node metric",
	}, []string{"name", "metric_type", "data_granularity", "export_timestamp"})

	nodeMetricDriftCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "node_metric_drift_total",
		Help:      "Total number of node metric drift",
	}, []string{"name", "data_granularity", "export_timestamp"})
)

type nodeMetric struct{}

func newNodeMetric() *nodeMetric {
	return &nodeMetric{}
}

func (nodeMetric *nodeMetric) setNodeMetricModelTime(
	name, dataGranularity, exportTimestamp string, val float64) {
	nodeModelTimeGauge.WithLabelValues(name, dataGranularity,
		exportTimestamp).Set(val)
}

func (nodeMetric *nodeMetric) addNodeMetricModelTimeTotal(
	name, dataGranularity, exportTimestamp string, val float64) {
	nodeModelTimeCounter.WithLabelValues(name, dataGranularity,
		exportTimestamp).Add(val)
}

func (nodeMetric *nodeMetric) setNodeMetricMAPE(name, metricType,
	dataGranularity, exportTimestamp string, val float64) {
	nodeMAPEGauge.WithLabelValues(name, metricType,
		dataGranularity, exportTimestamp).Set(val)
}

func (nodeMetric *nodeMetric) setNodeMetricRMSE(name, metricType,
	dataGranularity, exportTimestamp string, val float64) {
	nodeRMSEGauge.WithLabelValues(name, metricType,
		dataGranularity, exportTimestamp).Set(val)
}

func (nodeMetric *nodeMetric) addNodeMetricDrift(name,
	dataGranularity, exportTimestamp string, val float64) {
	nodeMetricDriftCounter.WithLabelValues(name, dataGranularity,
		exportTimestamp).Add(val)
}
