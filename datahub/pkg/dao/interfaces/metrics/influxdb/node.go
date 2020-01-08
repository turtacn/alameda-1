package influxdb

import (
	"context"

	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	RepoInfluxMetric "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/metrics"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	Utils "github.com/containers-ai/alameda/datahub/pkg/utils"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

type NodeMetrics struct {
	InfluxDBConfig InternalInflux.Config
}

func NewNodeMetricsWithConfig(config InternalInflux.Config) DaoMetricTypes.NodeMetricsDAO {
	return &NodeMetrics{InfluxDBConfig: config}
}

func (p *NodeMetrics) CreateMetrics(ctx context.Context, metrics DaoMetricTypes.NodeMetricMap) error {
	// Write node cpu metrics
	nodeCpuRepo := RepoInfluxMetric.NewNodeCpuRepositoryWithConfig(p.InfluxDBConfig)
	err := nodeCpuRepo.CreateMetrics(metrics.GetSamples(FormatEnum.MetricTypeCPUUsageSecondsPercentage))
	if err != nil {
		scope.Error(err.Error())
		return err
	}

	// Write node memory metrics
	nodeMemoryRepo := RepoInfluxMetric.NewNodeMemoryRepositoryWithConfig(p.InfluxDBConfig)
	err = nodeMemoryRepo.CreateMetrics(metrics.GetSamples(FormatEnum.MetricTypeMemoryUsageBytes))
	if err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}

func (p *NodeMetrics) ListMetrics(ctx context.Context, req DaoMetricTypes.ListNodeMetricsRequest) (DaoMetricTypes.NodeMetricMap, error) {
	nodeMetricMap := DaoMetricTypes.NewNodeMetricMap()

	// Read node cpu metrics
	if Utils.SliceContains(req.MetricTypes, FormatEnum.MetricTypeCPUUsageSecondsPercentage) {
		nodeCpuRepo := RepoInfluxMetric.NewNodeCpuRepositoryWithConfig(p.InfluxDBConfig)
		cpuMetrics, err := nodeCpuRepo.ListMetrics(req)
		if err != nil {
			scope.Error(err.Error())
			return DaoMetricTypes.NewNodeMetricMap(), err
		}
		for _, nodeMetric := range cpuMetrics {
			nodeMetricMap.AddNodeMetric(nodeMetric)
		}
	}

	// Read node memory metrics
	if Utils.SliceContains(req.MetricTypes, FormatEnum.MetricTypeMemoryUsageBytes) {
		nodeMemoryRepo := RepoInfluxMetric.NewNodeMemoryRepositoryWithConfig(p.InfluxDBConfig)
		memoryMetrics, err := nodeMemoryRepo.ListMetrics(req)
		if err != nil {
			scope.Error(err.Error())
			return DaoMetricTypes.NewNodeMetricMap(), err
		}
		for _, nodeMetric := range memoryMetrics {
			nodeMetricMap.AddNodeMetric(nodeMetric)
		}
	}

	return nodeMetricMap, nil
}
