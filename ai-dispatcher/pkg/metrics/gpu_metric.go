package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	gpuModelTimeGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "gpu_model_seconds",
		Help:      "Target modeling time of gpu",
	}, []string{"host", "minor_number", "data_granularity"})

	gpuModelTimeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "gpu_model_seconds_total",
		Help:      "Total target modeling time of gpu",
	}, []string{"host", "minor_number", "data_granularity"})

	gpuMAPEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "gpu_metric_mape",
		Help:      "MAPE of gpu metric",
	}, []string{"host", "minor_number", "metric_type", "data_granularity"})

	gpuMetricDriftCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "gpu_metric_drift_total",
		Help:      "Total number of gpu metric drift",
	}, []string{"host", "minor_number", "data_granularity"})
)

type gpuMetric struct{}

func newGPUMetric() *gpuMetric {
	return &gpuMetric{}
}

func (gpuMetric *gpuMetric) setGPUMetricModelTime(
	host, minor_number, dataGranularity string, val float64) {
	gpuModelTimeGauge.WithLabelValues(host,
		minor_number, dataGranularity).Set(val)
}

func (gpuMetric *gpuMetric) addGPUMetricModelTimeTotal(
	host, minor_number, dataGranularity string, val float64) {
	gpuModelTimeCounter.WithLabelValues(host,
		minor_number, dataGranularity).Add(val)
}

func (gpuMetric *gpuMetric) setGPUMetricMAPE(
	host, minor_number, metricType, dataGranularity string, val float64) {
	gpuMAPEGauge.WithLabelValues(host,
		minor_number, metricType, dataGranularity).Set(val)
}

func (gpuMetric *gpuMetric) addGPUMetricDrift(
	host, minor_number, dataGranularity string, val float64) {
	gpuMetricDriftCounter.WithLabelValues(host,
		minor_number, dataGranularity).Add(val)
}
