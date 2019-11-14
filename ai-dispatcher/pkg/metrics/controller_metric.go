package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	controllerModelTimeGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "controller_model_seconds",
		Help:      "Target modeling time of controller",
	}, []string{"namespace", "name", "kind", "data_granularity"})

	controllerModelTimeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "controller_model_seconds_total",
		Help:      "Total target modeling time of controller",
	}, []string{"namespace", "name", "kind", "data_granularity"})

	controllerMAPEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "controller_metric_mape",
		Help:      "MAPE of controller metric",
	}, []string{"namespace", "name", "kind", "metric_type", "data_granularity"})

	controllerRMSEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "controller_metric_rmse",
		Help:      "RMSE of controller metric",
	}, []string{"namespace", "name", "kind", "metric_type", "data_granularity"})

	controllerMetricDriftCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "controller_metric_drift_total",
		Help:      "Total number of controller metric drift",
	}, []string{"namespace", "name", "kind", "data_granularity"})
)

type controllerMetric struct{}

func newControllerMetric() *controllerMetric {
	return &controllerMetric{}
}

func (controllerMetric *controllerMetric) setControllerMetricModelTime(
	namespace, name, kind, dataGranularity string, val float64) {
	controllerModelTimeGauge.WithLabelValues(namespace,
		name, kind, dataGranularity).Set(val)
}

func (controllerMetric *controllerMetric) addControllerMetricModelTimeTotal(
	namespace, name, kind, dataGranularity string, val float64) {
	controllerModelTimeCounter.WithLabelValues(namespace,
		name, kind, dataGranularity).Add(val)
}

func (controllerMetric *controllerMetric) setControllerMetricMAPE(
	namespace, name, kind, metricType, dataGranularity string, val float64) {
	controllerMAPEGauge.WithLabelValues(namespace,
		name, kind, metricType, dataGranularity).Set(val)
}

func (controllerMetric *controllerMetric) setControllerMetricRMSE(
	namespace, name, kind, metricType, dataGranularity string, val float64) {
	controllerRMSEGauge.WithLabelValues(namespace, kind,
		name, metricType, dataGranularity).Set(val)
}

func (controllerMetric *controllerMetric) addControllerMetricDrift(
	namespace, name, kind, dataGranularity string, val float64) {
	controllerMetricDriftCounter.WithLabelValues(namespace,
		name, kind, dataGranularity).Add(val)
}
