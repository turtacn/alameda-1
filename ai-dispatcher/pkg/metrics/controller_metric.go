package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	controllerModelTimeGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "controller_model_seconds",
		Help:      "Target modeling time of controller metric",
	}, []string{"cluster_name", "namespace", "name", "kind", "data_granularity", "metric_type", "export_timestamp"})

	controllerModelTimeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "controller_model_seconds_total",
		Help:      "Total target modeling time of controller metric",
	}, []string{"cluster_name", "namespace", "name", "kind", "data_granularity", "metric_type", "export_timestamp"})

	controllerMAPEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "controller_metric_mape",
		Help:      "MAPE of controller metric",
	}, []string{"cluster_name", "namespace", "name", "kind", "data_granularity", "metric_type", "export_timestamp"})

	controllerRMSEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "controller_metric_rmse",
		Help:      "RMSE of controller metric",
	}, []string{"cluster_name", "namespace", "name", "kind", "data_granularity", "metric_type", "export_timestamp"})

	controllerMetricDriftCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "controller_metric_drift_total",
		Help:      "Total number of controller metric drift",
	}, []string{"cluster_name", "namespace", "name", "kind", "data_granularity", "metric_type", "export_timestamp"})
)

type controllerMetric struct{}

func newControllerMetric() *controllerMetric {
	return &controllerMetric{}
}

func (controllerMetric *controllerMetric) setControllerMetricModelTime(clusterID,
	namespace, name, kind, dataGranularity, metricType, exportTimestamp string,
	val float64) {
	controllerModelTimeGauge.WithLabelValues(clusterID, namespace,
		name, kind, dataGranularity, metricType, exportTimestamp).Set(val)
}

func (controllerMetric *controllerMetric) addControllerMetricModelTimeTotal(clusterID,
	namespace, name, kind, dataGranularity, metricType, exportTimestamp string,
	val float64) {
	controllerModelTimeCounter.WithLabelValues(clusterID, namespace,
		name, kind, dataGranularity, metricType, exportTimestamp).Add(val)
}

func (controllerMetric *controllerMetric) setControllerMetricMAPE(clusterID,
	namespace, name, kind, dataGranularity, metricType, exportTimestamp string,
	val float64) {
	controllerMAPEGauge.WithLabelValues(clusterID, namespace,
		name, kind, dataGranularity, metricType, exportTimestamp).Set(val)
}

func (controllerMetric *controllerMetric) setControllerMetricRMSE(clusterID,
	namespace, name, kind, dataGranularity, metricType, exportTimestamp string,
	val float64) {
	controllerRMSEGauge.WithLabelValues(clusterID, namespace, kind,
		name, dataGranularity, metricType, exportTimestamp).Set(val)
}

func (controllerMetric *controllerMetric) addControllerMetricDrift(clusterID,
	namespace, name, kind, dataGranularity, metricType, exportTimestamp string,
	val float64) {
	controllerMetricDriftCounter.WithLabelValues(clusterID, namespace,
		name, kind, dataGranularity, metricType, exportTimestamp).Add(val)
}
