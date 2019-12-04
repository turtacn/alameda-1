package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	podModelTimeGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "pod_model_seconds",
		Help:      "Target modeling time of pod",
	}, []string{"namespace", "name", "data_granularity", "export_timestamp"})

	podModelTimeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "pod_model_seconds_total",
		Help:      "Total target modeling time of pod",
	}, []string{"namespace", "name", "data_granularity", "export_timestamp"})

	containerMetricMAPEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "container_metric_mape",
		Help:      "MAPE of container metric",
	}, []string{"pod_namespace", "pod_name", "name", "metric_type", "data_granularity", "export_timestamp"})

	containerMetricRMSEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "container_metric_rmse",
		Help:      "RMSE of container metric",
	}, []string{"pod_namespace", "pod_name", "name", "metric_type", "data_granularity", "export_timestamp"})

	podMetricDriftCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "pod_metric_drift_total",
		Help:      "Total number of pod metric drift",
	}, []string{"namespace", "name", "data_granularity", "export_timestamp"})
)

type podMetric struct{}

func newPodMetric() *podMetric {
	return &podMetric{}
}

func (podMetric *podMetric) setPodMetricModelTime(podNS, podName,
	dataGranularity, exportTimestamp string, val float64) {
	podModelTimeGauge.WithLabelValues(podNS, podName,
		dataGranularity, exportTimestamp).Set(val)
}

func (podMetric *podMetric) addPodMetricModelTimeTotal(podNS, podName,
	dataGranularity, exportTimestamp string, val float64) {
		podModelTimeCounter.WithLabelValues(podNS, podName,
	dataGranularity, exportTimestamp).Add(val)
}

func (podMetric *podMetric) setContainerMetricMAPE(podNS, podName,
	name, metricType, dataGranularity, exportTimestamp string,
	val float64) {
	containerMetricMAPEGauge.WithLabelValues(podNS, podName,
		name, metricType, dataGranularity, exportTimestamp).Set(val)
}

func (podMetric *podMetric) setContainerMetricRMSE(podNS, podName,
	name, metricType, dataGranularity, exportTimestamp string,
	val float64) {
	containerMetricRMSEGauge.WithLabelValues(podNS, podName,
		name, metricType, dataGranularity, exportTimestamp).Set(val)
}

func (podMetric *podMetric) addPodMetricDrift(podNS, podName,
	dataGranularity, exportTimestamp string, val float64) {
	podMetricDriftCounter.WithLabelValues(podNS, podName,
		dataGranularity, exportTimestamp).Add(val)
}
