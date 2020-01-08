package influxdb

import (
	"context"

	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	RepoInfluxMetric "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/metrics"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	Utils "github.com/containers-ai/alameda/datahub/pkg/utils"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"github.com/pkg/errors"
)

type ClusterMetrics struct {
	InfluxDBConfig InternalInflux.Config
}

func NewClusterMetricsWithConfig(config InternalInflux.Config) DaoMetricTypes.ClusterMetricsDAO {
	return &ClusterMetrics{InfluxDBConfig: config}
}

func (n ClusterMetrics) CreateMetrics(ctx context.Context, m DaoMetricTypes.ClusterMetricMap) error {
	// Write cluster cpu metrics
	cpuRepo := RepoInfluxMetric.NewClusterCPURepositoryWithConfig(n.InfluxDBConfig)
	err := cpuRepo.CreateMetrics(ctx, m.GetSamples(FormatEnum.MetricTypeCPUUsageSecondsPercentage))
	if err != nil {
		scope.Error(err.Error())
		return errors.Wrap(err, "create application cpu metrics failed")
	}

	// Write cluster memory metrics
	memoryRepo := RepoInfluxMetric.NewClusterMemoryRepositoryWithConfig(n.InfluxDBConfig)
	err = memoryRepo.CreateMetrics(ctx, m.GetSamples(FormatEnum.MetricTypeMemoryUsageBytes))
	if err != nil {
		scope.Error(err.Error())
		return errors.Wrap(err, "create application memory metrics failed")
	}
	return nil
}

func (n ClusterMetrics) ListMetrics(ctx context.Context, req DaoMetricTypes.ListClusterMetricsRequest) (DaoMetricTypes.ClusterMetricMap, error) {
	metricMap := DaoMetricTypes.NewClusterMetricMap()

	// Read cluster cpu metrics
	if Utils.SliceContains(req.MetricTypes, FormatEnum.MetricTypeCPUUsageSecondsPercentage) {
		cpuRepo := RepoInfluxMetric.NewClusterCPURepositoryWithConfig(n.InfluxDBConfig)
		cpuMetricMap, err := cpuRepo.GetClusterMetricMap(ctx, req)
		if err != nil {
			scope.Error(err.Error())
			return metricMap, errors.Wrap(err, "get cluster cpu usage metric map failed")
		}
		for _, m := range cpuMetricMap.MetricMap {
			copyM := m
			metricMap.AddClusterMetric(copyM)
		}
	}

	// Read cluster memory metrics
	if Utils.SliceContains(req.MetricTypes, FormatEnum.MetricTypeMemoryUsageBytes) {
		memoryRepo := RepoInfluxMetric.NewClusterMemoryRepositoryWithConfig(n.InfluxDBConfig)
		memoryMetricMap, err := memoryRepo.GetClusterMetricMap(ctx, req)
		if err != nil {
			scope.Error(err.Error())
			return metricMap, errors.Wrap(err, "get cluster memory usage metric map failed")
		}
		for _, m := range memoryMetricMap.MetricMap {
			copyM := m
			metricMap.AddClusterMetric(copyM)
		}
	}

	return metricMap, nil
}
