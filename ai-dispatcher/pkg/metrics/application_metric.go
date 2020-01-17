package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	applicationModelTimeGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "application_model_seconds",
		Help:      "Target modeling time of application metric",
	}, []string{"cluster_name", "namespace", "name", "data_granularity", "metric_type", "export_timestamp"})

	applicationModelTimeCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "application_model_seconds_total",
		Help:      "Total target modeling time of application metric",
	}, []string{"cluster_name", "namespace", "name", "data_granularity", "metric_type", "export_timestamp"})

	applicationMAPEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "application_metric_mape",
		Help:      "MAPE of application metric",
	}, []string{"cluster_name", "namespace", "name", "data_granularity", "metric_type", "export_timestamp"})

	applicationRMSEGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "application_metric_rmse",
		Help:      "RMSE of application metric",
	}, []string{"cluster_name", "namespace", "name", "data_granularity", "metric_type", "export_timestamp"})

	applicationMetricDriftCounter = promauto.NewCounterVec(prometheus.CounterOpts{
		Subsystem: "alameda_ai_dispatcher",
		Name:      "application_metric_drift_total",
		Help:      "Total number of application metric drift",
	}, []string{"cluster_name", "namespace", "name", "data_granularity", "metric_type", "export_timestamp"})
)

type applicationMetric struct{}

func newApplicationMetric() *applicationMetric {
	return &applicationMetric{}
}

func (applicationMetric *applicationMetric) setApplicationMetricModelTime(clusterID,
	appNS, appName, dataGranularity, metricType, exportTimestamp string, val float64) {
	applicationModelTimeGauge.WithLabelValues(clusterID, appNS, appName,
		dataGranularity, metricType, exportTimestamp).Set(val)
}

func (applicationMetric *applicationMetric) addApplicationMetricModelTimeTotal(clusterID,
	appNS, appName, dataGranularity, metricType, exportTimestamp string, val float64) {
	applicationModelTimeCounter.WithLabelValues(clusterID, appNS,
		appName, dataGranularity, metricType, exportTimestamp).Add(val)
}

func (applicationMetric *applicationMetric) setApplicationMetricMAPE(clusterID,
	appNS, appName, dataGranularity, metricType, exportTimestamp string,
	val float64) {
	applicationMAPEGauge.WithLabelValues(clusterID, appNS, appName,
		dataGranularity, metricType, exportTimestamp).Set(val)
}

func (applicationMetric *applicationMetric) setApplicationMetricRMSE(clusterID,
	appNS, appName, dataGranularity, metricType, exportTimestamp string,
	val float64) {
	applicationRMSEGauge.WithLabelValues(clusterID, appNS, appName,
		dataGranularity, metricType, exportTimestamp).Set(val)
}

func (applicationMetric *applicationMetric) addApplicationMetricDrift(clusterID,
	appNS, appName, dataGranularity, metricType, exportTimestamp string,
	val float64) {
	applicationMetricDriftCounter.WithLabelValues(clusterID, appNS,
		appName, dataGranularity, metricType, exportTimestamp).Add(val)
}
