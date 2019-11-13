package metrics

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/prometheus"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
)

func NewClusterMetricsReaderDAO(config config.Config) types.ClusterMetricsDAO {
	switch config.Apis.Metrics.Source {
	case "influxdb":
		return influxdb.NewClusterMetricsWithConfig(*config.InfluxDB)
	case "prometheus":
		return prometheus.NewClusterMetricsWithConfig(*config.Prometheus, *config.InfluxDB, config.ClusterUID)
	default:
		return prometheus.NewClusterMetricsWithConfig(*config.Prometheus, *config.InfluxDB, config.ClusterUID)
	}
}

func NewClusterMetricsWriterDAO(config config.Config) types.ClusterMetricsDAO {
	switch config.Apis.Metrics.Target {
	case "influxdb":
		return influxdb.NewClusterMetricsWithConfig(*config.InfluxDB)
	case "prometheus":
		return prometheus.NewClusterMetricsWithConfig(*config.Prometheus, *config.InfluxDB, config.ClusterUID)
	default:
		return prometheus.NewClusterMetricsWithConfig(*config.Prometheus, *config.InfluxDB, config.ClusterUID)
	}
}
