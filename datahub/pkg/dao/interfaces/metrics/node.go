package metrics

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/prometheus"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
)

func NewNodeMetricsReaderDAO(config config.Config) types.NodeMetricsDAO {
	switch config.Apis.Metrics.Source {
	case "influxdb":
		return influxdb.NewNodeMetricsWithConfig(*config.InfluxDB)
	case "prometheus":
		return prometheus.NewNodeMetricsWithConfig(*config.Prometheus)
	default:
		return prometheus.NewNodeMetricsWithConfig(*config.Prometheus)
	}
}

func NewNodeMetricsWriterDAO(config config.Config) types.NodeMetricsDAO {
	switch config.Apis.Metrics.Target {
	case "influxdb":
		return influxdb.NewNodeMetricsWithConfig(*config.InfluxDB)
	case "prometheus":
		return prometheus.NewNodeMetricsWithConfig(*config.Prometheus)
	default:
		return prometheus.NewNodeMetricsWithConfig(*config.Prometheus)
	}
}
