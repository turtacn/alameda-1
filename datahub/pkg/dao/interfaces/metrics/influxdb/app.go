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

type AppMetrics struct {
	InfluxDBConfig InternalInflux.Config
}

func NewAppMetricsWithConfig(config InternalInflux.Config) DaoMetricTypes.AppMetricsDAO {
	return &AppMetrics{InfluxDBConfig: config}
}

func (n AppMetrics) CreateMetrics(ctx context.Context, m DaoMetricTypes.AppMetricMap) error {
	// Write app cpu metrics
	appCPURepo := RepoInfluxMetric.NewApplicationCPURepositoryWithConfig(n.InfluxDBConfig)
	err := appCPURepo.CreateMetrics(ctx, m.GetSamples(FormatEnum.MetricTypeCPUUsageSecondsPercentage))
	if err != nil {
		scope.Error(err.Error())
		return errors.Wrap(err, "create application cpu metrics failed")
	}

	// Write app memory metrics
	appMemoryRepo := RepoInfluxMetric.NewApplicationMemoryRepositoryWithConfig(n.InfluxDBConfig)
	err = appMemoryRepo.CreateMetrics(ctx, m.GetSamples(FormatEnum.MetricTypeMemoryUsageBytes))
	if err != nil {
		scope.Error(err.Error())
		return errors.Wrap(err, "create application memory metrics failed")
	}
	return nil
}

func (n AppMetrics) ListMetrics(ctx context.Context, req DaoMetricTypes.ListAppMetricsRequest) (DaoMetricTypes.AppMetricMap, error) {
	metricMap := DaoMetricTypes.NewAppMetricMap()

	// Read app cpu metrics
	if Utils.SliceContains(req.MetricTypes, FormatEnum.MetricTypeCPUUsageSecondsPercentage) {
		appCPURepo := RepoInfluxMetric.NewApplicationCPURepositoryWithConfig(n.InfluxDBConfig)
		cpuMetricMap, err := appCPURepo.GetApplicationMetricMap(ctx, req)
		if err != nil {
			scope.Error(err.Error())
			return metricMap, errors.Wrap(err, "get application cpu usage metric map failed")
		}
		for _, m := range cpuMetricMap.MetricMap {
			copyM := m
			metricMap.AddAppMetric(copyM)
		}
	}

	// Read app memory metrics
	if Utils.SliceContains(req.MetricTypes, FormatEnum.MetricTypeMemoryUsageBytes) {
		appMemoryRepo := RepoInfluxMetric.NewApplicationMemoryRepositoryWithConfig(n.InfluxDBConfig)
		memoryMetricMap, err := appMemoryRepo.GetApplicationMetricMap(ctx, req)
		if err != nil {
			scope.Error(err.Error())
			return metricMap, errors.Wrap(err, "get application memory usage metric map failed")
		}
		for _, m := range memoryMetricMap.MetricMap {
			copyM := m
			metricMap.AddAppMetric(copyM)
		}
	}

	return metricMap, nil
}
