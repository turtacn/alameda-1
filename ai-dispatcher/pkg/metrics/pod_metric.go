package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ctMetricModelTimeGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "container_metric_model_seconds",
		Help:      "Target modeling time of container metric",
	}, []string{"cluster_name", "pod_namespace", "pod_name", "name", "data_granularity", "metric_type", "export_timestamp"})

	ctMetricModelTimeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "container_metric_model_seconds_total",
		Help:      "Total target modeling time of container metric",
	}, []string{"cluster_name", "pod_namespace", "pod_name", "name", "data_granularity", "metric_type", "export_timestamp"})

	containerMetricMAPEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "container_metric_mape",
		Help:      "MAPE of container metric",
	}, []string{"cluster_name", "pod_namespace", "pod_name", "name", "data_granularity", "metric_type", "export_timestamp"})

	containerMetricRMSEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "container_metric_rmse",
		Help:      "RMSE of container metric",
	}, []string{"cluster_name", "pod_namespace", "pod_name", "name", "data_granularity", "metric_type", "export_timestamp"})

	containerMetricDriftCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "container_metric_drift_total",
		Help:      "Total number of container metric drift",
	}, []string{"cluster_name", "pod_namespace", "pod_name", "name", "data_granularity", "metric_type", "export_timestamp"})
)

type podMetric struct{}

func newPodMetric() *podMetric {
	return &podMetric{}
}

func (podMetric *podMetric) setContainerMetricModelTime(clusterID, podNS, podName, ctName,
	dataGranularity, metricType, exportTimestamp string, val float64) {
	ctMetricModelTimeGauge.WithLabelValues(clusterID, podNS, podName, ctName,
		dataGranularity, metricType, exportTimestamp).Set(val)
}

func (podMetric *podMetric) addContainerMetricModelTimeTotal(clusterID, podNS, podName, ctName,
	dataGranularity, metricType, exportTimestamp string, val float64) {
	ctMetricModelTimeCounter.WithLabelValues(clusterID, podNS, podName, ctName,
		dataGranularity, metricType, exportTimestamp).Add(val)
}

func (podMetric *podMetric) setContainerMetricMAPE(clusterID, podNS, podName,
	name, dataGranularity, metricType, exportTimestamp string,
	val float64) {
	containerMetricMAPEGauge.WithLabelValues(clusterID, podNS, podName,
		name, dataGranularity, metricType, exportTimestamp).Set(val)
}

func (podMetric *podMetric) setContainerMetricRMSE(clusterID, podNS, podName,
	name, dataGranularity, metricType, exportTimestamp string,
	val float64) {
	containerMetricRMSEGauge.WithLabelValues(clusterID, podNS, podName,
		name, dataGranularity, metricType, exportTimestamp).Set(val)
}

func (podMetric *podMetric) addPodMetricDrift(clusterID, podNS, podName, name,
	dataGranularity, metricType, exportTimestamp string, val float64) {
	containerMetricDriftCounter.WithLabelValues(clusterID, podNS, podName, name,
		dataGranularity, metricType, exportTimestamp).Add(val)
}
