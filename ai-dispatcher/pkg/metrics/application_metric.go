package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	applicationModelTimeGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "application_model_seconds",
		Help:      "Target modeling time of application",
	}, []string{"namespace", "name", "data_granularity", "export_timestamp"})

	applicationModelTimeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "application_model_seconds_total",
		Help:      "Total target modeling time of application",
	}, []string{"namespace", "name", "data_granularity", "export_timestamp"})

	applicationMAPEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "application_metric_mape",
		Help:      "MAPE of application metric",
	}, []string{"namespace", "name", "metric_type", "data_granularity", "export_timestamp"})

	applicationRMSEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "application_metric_rmse",
		Help:      "RMSE of application metric",
	}, []string{"namespace", "name", "metric_type", "data_granularity", "export_timestamp"})

	applicationMetricDriftCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "application_metric_drift_total",
		Help:      "Total number of application metric drift",
	}, []string{"namespace", "name", "data_granularity", "export_timestamp"})
)

type applicationMetric struct{}

func newApplicationMetric() *applicationMetric {
	return &applicationMetric{}
}

func (applicationMetric *applicationMetric) setApplicationMetricModelTime(
	host, minor_number, dataGranularity, exportTimestamp string, val float64) {
	applicationModelTimeGauge.WithLabelValues(host, minor_number,
		dataGranularity, exportTimestamp).Set(val)
}

func (applicationMetric *applicationMetric) addApplicationMetricModelTimeTotal(
	host, minor_number, dataGranularity, exportTimestamp string, val float64) {
	applicationModelTimeCounter.WithLabelValues(host,
		minor_number, dataGranularity, exportTimestamp).Add(val)
}

func (applicationMetric *applicationMetric) setApplicationMetricMAPE(
	host, minor_number, metricType, dataGranularity, exportTimestamp string,
	val float64) {
	applicationMAPEGauge.WithLabelValues(host, minor_number,
		metricType, dataGranularity, exportTimestamp).Set(val)
}

func (applicationMetric *applicationMetric) setApplicationMetricRMSE(
	host, minor_number, metricType, dataGranularity, exportTimestamp string,
	val float64) {
	applicationRMSEGauge.WithLabelValues(host, minor_number,
		metricType, dataGranularity, exportTimestamp).Set(val)
}

func (applicationMetric *applicationMetric) addApplicationMetricDrift(
	host, minor_number, dataGranularity, exportTimestamp string,
	val float64) {
	applicationMetricDriftCounter.WithLabelValues(host,
		minor_number, dataGranularity, exportTimestamp).Add(val)
}
