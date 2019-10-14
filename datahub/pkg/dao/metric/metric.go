package metric

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/metric/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/metric/prometheus"
	"github.com/containers-ai/alameda/datahub/pkg/dao/metric/types"
)

func NewNodeMetricsDAO(config config.Config) types.NodeMetricsDAO {
	switch config.Apis.Metrics.Source {
	case "influxdb":
		return influxdb.NewNodeMetricsWithConfig(*config.InfluxDB)
	case "prometheus":
		return prometheus.NewNodeMetricsWithConfig(*config.Prometheus)
	default:
		return prometheus.NewNodeMetricsWithConfig(*config.Prometheus)
	}
}

func NewPodMetricsDAO(config config.Config) types.PodMetricsDAO {
	switch config.Apis.Metrics.Source {
	case "influxdb":
		return influxdb.NewPodMetricsWithConfig(*config.InfluxDB)
	case "prometheus":
		return prometheus.NewPodMetricsWithConfig(*config.Prometheus)
	default:
		return prometheus.NewPodMetricsWithConfig(*config.Prometheus)
	}
}
