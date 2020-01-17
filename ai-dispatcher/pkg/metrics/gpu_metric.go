package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	gpuModelTimeGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "gpu_model_seconds",
		Help:      "Target modeling time of gpu metric",
	}, []string{"cluster_name", "host", "minor_number", "data_granularity", "metric_type", "export_timestamp"})

	gpuModelTimeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "gpu_model_seconds_total",
		Help:      "Total target modeling time of gpu metric",
	}, []string{"cluster_name", "host", "minor_number", "data_granularity", "metric_type", "export_timestamp"})

	gpuMAPEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "gpu_metric_mape",
		Help:      "MAPE of gpu metric",
	}, []string{"cluster_name", "host", "minor_number", "data_granularity", "metric_type", "export_timestamp"})

	gpuRMSEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "gpu_metric_rmse",
		Help:      "RMSE of gpu metric",
	}, []string{"cluster_name", "host", "minor_number", "data_granularity", "metric_type", "export_timestamp"})

	gpuMetricDriftCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "gpu_metric_drift_total",
		Help:      "Total number of gpu metric drift",
	}, []string{"cluster_name", "host", "minor_number", "data_granularity", "metric_type", "export_timestamp"})
)

type gpuMetric struct{}

func newGPUMetric() *gpuMetric {
	return &gpuMetric{}
}

func (gpuMetric *gpuMetric) setGPUMetricModelTime(clusterID, host, minorNumber,
	dataGranularity, metricType, exportTimestamp string, val float64) {
	gpuModelTimeGauge.WithLabelValues(clusterID, host, minorNumber,
		dataGranularity, metricType, exportTimestamp).Set(val)
}

func (gpuMetric *gpuMetric) addGPUMetricModelTimeTotal(clusterID, host,
	minorNumber, dataGranularity, metricType, exportTimestamp string, val float64) {
	gpuModelTimeCounter.WithLabelValues(clusterID, host,
		minorNumber, dataGranularity, metricType, exportTimestamp).Add(val)
}

func (gpuMetric *gpuMetric) setGPUMetricMAPE(clusterID, host, minorNumber,
	dataGranularity, metricType, exportTimestamp string,
	val float64) {
	gpuMAPEGauge.WithLabelValues(clusterID, host, minorNumber,
		dataGranularity, metricType, exportTimestamp).Set(val)
}

func (gpuMetric *gpuMetric) setGPUMetricRMSE(clusterID, host, minorNumber,
	dataGranularity, metricType, exportTimestamp string,
	val float64) {
	gpuRMSEGauge.WithLabelValues(clusterID, host, minorNumber,
		dataGranularity, metricType, exportTimestamp).Set(val)
}

func (gpuMetric *gpuMetric) addGPUMetricDrift(clusterID, host, minorNumber,
	dataGranularity, metricType, exportTimestamp string, val float64) {
	gpuMetricDriftCounter.WithLabelValues(clusterID, host,
		minorNumber, dataGranularity, metricType, exportTimestamp).Add(val)
}
