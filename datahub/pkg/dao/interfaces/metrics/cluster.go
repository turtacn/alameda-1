package metrics

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	DaoClusterStatus "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/prometheus"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
)

func NewClusterMetricsReaderDAO(config config.Config) types.ClusterMetricsDAO {
	switch config.Apis.Metrics.Source {
	case "influxdb":
		return influxdb.NewClusterMetricsWithConfig(*config.InfluxDB)
	case "prometheus":
		return prometheus.NewClusterMetricsWithConfig(*config.Prometheus, DaoClusterStatus.NewClusterDAO(config), DaoClusterStatus.NewNodeDAO(config), config.ClusterUID)
	default:
		return prometheus.NewClusterMetricsWithConfig(*config.Prometheus, DaoClusterStatus.NewClusterDAO(config), DaoClusterStatus.NewNodeDAO(config), config.ClusterUID)
	}
}

func NewClusterMetricsWriterDAO(config config.Config) types.ClusterMetricsDAO {
	switch config.Apis.Metrics.Target {
	case "influxdb":
		return influxdb.NewClusterMetricsWithConfig(*config.InfluxDB)
	case "prometheus":
		return prometheus.NewClusterMetricsWithConfig(*config.Prometheus, DaoClusterStatus.NewClusterDAO(config), DaoClusterStatus.NewNodeDAO(config), config.ClusterUID)
	default:
		return prometheus.NewClusterMetricsWithConfig(*config.Prometheus, DaoClusterStatus.NewClusterDAO(config), DaoClusterStatus.NewNodeDAO(config), config.ClusterUID)
	}
}
