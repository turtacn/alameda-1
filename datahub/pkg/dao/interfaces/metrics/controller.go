package metrics

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/prometheus"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
)

func NewControllerMetricsReaderDAO(config config.Config) types.ControllerMetricsDAO {
	switch config.Apis.Metrics.Source {
	case "influxdb":
		return influxdb.NewControllerMetricsWithConfig(*config.InfluxDB)
	case "prometheus":
		return prometheus.NewControllerMetricsWithConfig(*config.Prometheus, *config.InfluxDB, config.ClusterUID)
	default:
		return prometheus.NewControllerMetricsWithConfig(*config.Prometheus, *config.InfluxDB, config.ClusterUID)
	}
}

func NewControllerMetricsWriterDAO(config config.Config) types.ControllerMetricsDAO {
	switch config.Apis.Metrics.Target {
	case "influxdb":
		return influxdb.NewControllerMetricsWithConfig(*config.InfluxDB)
	case "prometheus":
		return prometheus.NewControllerMetricsWithConfig(*config.Prometheus, *config.InfluxDB, config.ClusterUID)
	default:
		return prometheus.NewControllerMetricsWithConfig(*config.Prometheus, *config.InfluxDB, config.ClusterUID)
	}
}
