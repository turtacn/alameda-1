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

type ControllerMetrics struct {
	InfluxDBConfig InternalInflux.Config
}

func NewControllerMetricsWithConfig(config InternalInflux.Config) DaoMetricTypes.ControllerMetricsDAO {
	return &ControllerMetrics{InfluxDBConfig: config}
}

func (n ControllerMetrics) CreateMetrics(ctx context.Context, m DaoMetricTypes.ControllerMetricMap) error {
	// Write controller cpu metrics
	controllerCPURepo := RepoInfluxMetric.NewControllerCPURepositoryWithConfig(n.InfluxDBConfig)
	err := controllerCPURepo.CreateMetrics(ctx, m.GetSamples(FormatEnum.MetricTypeCPUUsageSecondsPercentage))
	if err != nil {
		scope.Error(err.Error())
		return errors.Wrap(err, "create controller cpu metrics failed")
	}

	// Write controller memory metrics
	memoryRepo := RepoInfluxMetric.NewControllerMemoryRepositoryWithConfig(n.InfluxDBConfig)
	err = memoryRepo.CreateMetrics(ctx, m.GetSamples(FormatEnum.MetricTypeMemoryUsageBytes))
	if err != nil {
		scope.Error(err.Error())
		return errors.Wrap(err, "create controller memory metrics failed")
	}

	return nil
}

func (n ControllerMetrics) ListMetrics(ctx context.Context, req DaoMetricTypes.ListControllerMetricsRequest) (DaoMetricTypes.ControllerMetricMap, error) {
	metricMap := DaoMetricTypes.NewControllerMetricMap()

	// Read controller cpu metrics
	if Utils.SliceContains(req.MetricTypes, FormatEnum.MetricTypeCPUUsageSecondsPercentage) {
		controllerCPURepo := RepoInfluxMetric.NewControllerCPURepositoryWithConfig(n.InfluxDBConfig)
		cpuMetricMap, err := controllerCPURepo.GetControllerMetricMap(ctx, req)
		if err != nil {
			scope.Error(err.Error())
			return metricMap, errors.Wrap(err, "get controller cpu usage metric map failed")
		}
		for _, m := range cpuMetricMap.MetricMap {
			copyM := m
			metricMap.AddControllerMetric(copyM)
		}
	}

	// Read controller memory metrics
	if Utils.SliceContains(req.MetricTypes, FormatEnum.MetricTypeMemoryUsageBytes) {
		controllerMemoryRepo := RepoInfluxMetric.NewControllerMemoryRepositoryWithConfig(n.InfluxDBConfig)
		memoryMetricMap, err := controllerMemoryRepo.GetControllerMetricMap(ctx, req)
		if err != nil {
			scope.Error(err.Error())
			return metricMap, errors.Wrap(err, "get controller memory usage metric map failed")
		}
		for _, m := range memoryMetricMap.MetricMap {
			copyM := m
			metricMap.AddControllerMetric(copyM)
		}
	}

	return metricMap, nil
}
