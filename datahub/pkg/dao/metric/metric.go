package metric

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/metric/prometheus"
	"github.com/containers-ai/alameda/datahub/pkg/dao/metric/types"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
)

func NewNodeMetricsDAO(config InternalPromth.Config) types.NodeMetricsDAO {
	return prometheus.NewNodeMetricsWithConfig(config)
}

func NewPodMetricsDAO(config InternalPromth.Config) types.PodMetricsDAO {
	return prometheus.NewPodMetricsWithConfig(config)
}
