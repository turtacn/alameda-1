package metrics

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	DaoClusterStatus "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/prometheus"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
)

func NewAppMetricsReaderDAO(config config.Config) types.AppMetricsDAO {
	switch config.Apis.Metrics.Source {
	case "influxdb":
		return influxdb.NewAppMetricsWithConfig(*config.InfluxDB)
	case "prometheus":
		return prometheus.NewAppMetricsWithConfig(*config.Prometheus, DaoClusterStatus.NewApplicationDAO(config), config.ClusterUID)
	default:
		return prometheus.NewAppMetricsWithConfig(*config.Prometheus, DaoClusterStatus.NewApplicationDAO(config), config.ClusterUID)
	}
}

func NewAppMetricsWriterDAO(config config.Config) types.AppMetricsDAO {
	switch config.Apis.Metrics.Target {
	case "influxdb":
		return influxdb.NewAppMetricsWithConfig(*config.InfluxDB)
	case "prometheus":
		return prometheus.NewAppMetricsWithConfig(*config.Prometheus, DaoClusterStatus.NewApplicationDAO(config), config.ClusterUID)
	default:
		return prometheus.NewAppMetricsWithConfig(*config.Prometheus, DaoClusterStatus.NewApplicationDAO(config), config.ClusterUID)
	}
}
