package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	nodeModelTimeGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "node_model_seconds",
		Help:      "Target modeling time of node metric",
	}, []string{"cluster_name", "name", "data_granularity", "metric_type", "export_timestamp"})

	nodeModelTimeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "node_model_seconds_total",
		Help:      "Total target modeling time of node metric",
	}, []string{"cluster_name", "name", "data_granularity", "metric_type", "export_timestamp"})

	nodeMAPEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "node_metric_mape",
		Help:      "MAPE of node metric",
	}, []string{"cluster_name", "name", "data_granularity", "metric_type", "export_timestamp"})

	nodeRMSEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "node_metric_rmse",
		Help:      "RMSE of node metric",
	}, []string{"cluster_name", "name", "data_granularity", "metric_type", "export_timestamp"})

	nodeMetricDriftCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "node_metric_drift_total",
		Help:      "Total number of node metric drift",
	}, []string{"cluster_name", "name", "data_granularity", "metric_type", "export_timestamp"})
)

type nodeMetric struct{}

func newNodeMetric() *nodeMetric {
	return &nodeMetric{}
}

func (nodeMetric *nodeMetric) setNodeMetricModelTime(clusterID,
	name, dataGranularity, metricType, exportTimestamp string, val float64) {
	nodeModelTimeGauge.WithLabelValues(clusterID, name, dataGranularity, metricType,
		exportTimestamp).Set(val)
}

func (nodeMetric *nodeMetric) addNodeMetricModelTimeTotal(clusterID,
	name, dataGranularity, metricType, exportTimestamp string, val float64) {
	nodeModelTimeCounter.WithLabelValues(clusterID, name, dataGranularity, metricType,
		exportTimestamp).Add(val)
}

func (nodeMetric *nodeMetric) setNodeMetricMAPE(clusterID, name,
	dataGranularity, metricType, exportTimestamp string, val float64) {
	nodeMAPEGauge.WithLabelValues(clusterID, name,
		dataGranularity, metricType, exportTimestamp).Set(val)
}

func (nodeMetric *nodeMetric) setNodeMetricRMSE(clusterID, name,
	dataGranularity, metricType, exportTimestamp string, val float64) {
	nodeRMSEGauge.WithLabelValues(clusterID, name,
		dataGranularity, metricType, exportTimestamp).Set(val)
}

func (nodeMetric *nodeMetric) addNodeMetricDrift(clusterID, name,
	dataGranularity, metricType, exportTimestamp string, val float64) {
	nodeMetricDriftCounter.WithLabelValues(clusterID, name, dataGranularity, metricType,
		exportTimestamp).Add(val)
}
