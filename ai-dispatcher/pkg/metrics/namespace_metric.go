package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	namespaceModelTimeGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "namespace_model_seconds",
		Help:      "Target modeling time of namespace metric",
	}, []string{"cluster_name", "name", "data_granularity", "metric_type", "export_timestamp"})

	namespaceModelTimeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "namespace_model_seconds_total",
		Help:      "Total target modeling time of namespace metric",
	}, []string{"cluster_name", "name", "data_granularity", "metric_type", "export_timestamp"})

	namespaceMAPEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "namespace_metric_mape",
		Help:      "MAPE of namespace metric",
	}, []string{"cluster_name", "name", "data_granularity", "metric_type", "export_timestamp"})

	namespaceRMSEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "namespace_metric_rmse",
		Help:      "RMSE of namespace metric",
	}, []string{"cluster_name", "name", "data_granularity", "metric_type", "export_timestamp"})

	namespaceMetricDriftCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "namespace_metric_drift_total",
		Help:      "Total number of namespace metric drift",
	}, []string{"cluster_name", "name", "data_granularity", "metric_type", "export_timestamp"})
)

type namespaceMetric struct{}

func newNamespaceMetric() *namespaceMetric {
	return &namespaceMetric{}
}

func (namespaceMetric *namespaceMetric) setNamespaceMetricModelTime(clusterID,
	name, dataGranularity, metricType, exportTimestamp string, val float64) {
	namespaceModelTimeGauge.WithLabelValues(clusterID, name,
		dataGranularity, metricType, exportTimestamp).Set(val)
}

func (namespaceMetric *namespaceMetric) addNamespaceMetricModelTimeTotal(clusterID,
	name, dataGranularity, metricType, exportTimestamp string, val float64) {
	namespaceModelTimeCounter.WithLabelValues(clusterID, name,
		dataGranularity, metricType, exportTimestamp).Add(val)
}

func (namespaceMetric *namespaceMetric) setNamespaceMetricMAPE(clusterID,
	name, dataGranularity, metricType, exportTimestamp string, val float64) {
	namespaceMAPEGauge.WithLabelValues(clusterID, name,
		dataGranularity, metricType, exportTimestamp).Set(val)
}

func (namespaceMetric *namespaceMetric) setNamespaceMetricRMSE(clusterID,
	name, dataGranularity, metricType, exportTimestamp string, val float64) {
	namespaceRMSEGauge.WithLabelValues(clusterID, name,
		dataGranularity, metricType, exportTimestamp).Set(val)
}

func (namespaceMetric *namespaceMetric) addNamespaceMetricDrift(clusterID,
	name, dataGranularity, metricType, exportTimestamp string, val float64) {
	namespaceMetricDriftCounter.WithLabelValues(clusterID, name,
		dataGranularity, metricType, exportTimestamp).Add(val)
}
