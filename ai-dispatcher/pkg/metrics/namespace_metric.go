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
	}, []string{"name", "data_granularity"})

	namespaceModelTimeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "namespace_model_seconds_total",
		Help:      "Total target modeling time of namespace",
	}, []string{"name", "data_granularity"})

	namespaceMAPEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "namespace_metric_mape",
		Help:      "MAPE of namespace metric",
	}, []string{"name", "metric_type", "data_granularity"})

	namespaceRMSEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "namespace_metric_rmse",
		Help:      "RMSE of namespace metric",
	}, []string{"name", "metric_type", "data_granularity"})

	namespaceMetricDriftCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "namespace_metric_drift_total",
		Help:      "Total number of namespace metric drift",
	}, []string{"name", "data_granularity"})
)

type namespaceMetric struct{}

func newNamespaceMetric() *namespaceMetric {
	return &namespaceMetric{}
}

func (namespaceMetric *namespaceMetric) setNamespaceMetricModelTime(
	name, dataGranularity string, val float64) {
	namespaceModelTimeGauge.WithLabelValues(name, dataGranularity).Set(val)
}

func (namespaceMetric *namespaceMetric) addNamespaceMetricModelTimeTotal(
	name, dataGranularity string, val float64) {
	namespaceModelTimeCounter.WithLabelValues(name, dataGranularity).Add(val)
}

func (namespaceMetric *namespaceMetric) setNamespaceMetricMAPE(
	name, metricType, dataGranularity string, val float64) {
	namespaceMAPEGauge.WithLabelValues(name,
		metricType, dataGranularity).Set(val)
}

func (namespaceMetric *namespaceMetric) setNamespaceMetricRMSE(
	name, metricType, dataGranularity string, val float64) {
	namespaceRMSEGauge.WithLabelValues(name,
		metricType, dataGranularity).Set(val)
}

func (namespaceMetric *namespaceMetric) addNamespaceMetricDrift(
	name, dataGranularity string, val float64) {
	namespaceMetricDriftCounter.WithLabelValues(name, dataGranularity).Add(val)
}
