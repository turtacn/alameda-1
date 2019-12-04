package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	namespaceModelTimeGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "namespace_model_seconds",
		Help:      "Target modeling time of namespace",
	}, []string{"name", "data_granularity", "export_timestamp"})

	namespaceModelTimeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "namespace_model_seconds_total",
		Help:      "Total target modeling time of namespace",
	}, []string{"name", "data_granularity", "export_timestamp"})

	namespaceMAPEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "namespace_metric_mape",
		Help:      "MAPE of namespace metric",
	}, []string{"name", "metric_type", "data_granularity", "export_timestamp"})

	namespaceRMSEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "namespace_metric_rmse",
		Help:      "RMSE of namespace metric",
	}, []string{"name", "metric_type", "data_granularity", "export_timestamp"})

	namespaceMetricDriftCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "namespace_metric_drift_total",
		Help:      "Total number of namespace metric drift",
	}, []string{"name", "data_granularity", "export_timestamp"})
)

type namespaceMetric struct{}

func newNamespaceMetric() *namespaceMetric {
	return &namespaceMetric{}
}

func (namespaceMetric *namespaceMetric) setNamespaceMetricModelTime(
	name, dataGranularity, exportTimestamp string, val float64) {
	namespaceModelTimeGauge.WithLabelValues(name,
		dataGranularity, exportTimestamp).Set(val)
}

func (namespaceMetric *namespaceMetric) addNamespaceMetricModelTimeTotal(
	name, dataGranularity, exportTimestamp string, val float64) {
	namespaceModelTimeCounter.WithLabelValues(name,
		dataGranularity, exportTimestamp).Add(val)
}

func (namespaceMetric *namespaceMetric) setNamespaceMetricMAPE(
	name, metricType, dataGranularity, exportTimestamp string, val float64) {
	namespaceMAPEGauge.WithLabelValues(name,
		metricType, dataGranularity, exportTimestamp).Set(val)
}

func (namespaceMetric *namespaceMetric) setNamespaceMetricRMSE(
	name, metricType, dataGranularity, exportTimestamp string, val float64) {
	namespaceRMSEGauge.WithLabelValues(name,
		metricType, dataGranularity, exportTimestamp).Set(val)
}

func (namespaceMetric *namespaceMetric) addNamespaceMetricDrift(
	name, dataGranularity, exportTimestamp string, val float64) {
	namespaceMetricDriftCounter.WithLabelValues(name,
		dataGranularity, exportTimestamp).Add(val)
}
