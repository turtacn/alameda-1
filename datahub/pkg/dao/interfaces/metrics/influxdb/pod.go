package influxdb

import (
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	RepoInfluxMetric "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/metrics"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

type PodMetrics struct {
	InfluxDBConfig InternalInflux.Config
}

func NewPodMetricsWithConfig(config InternalInflux.Config) DaoMetricTypes.PodMetricsDAO {
	return &PodMetrics{InfluxDBConfig: config}
}

func (p *PodMetrics) CreateMetrics(metrics DaoMetricTypes.PodMetricMap) error {
	// Write container cpu metrics
	containerCpuRepo := RepoInfluxMetric.NewContainerCpuRepositoryWithConfig(p.InfluxDBConfig)
	cpuSampleList := make([]*DaoMetricTypes.ContainerMetricSample, 0)
	for _, podMetric := range metrics.MetricMap {
		for _, containerMetric := range podMetric.ContainerMetricMap.MetricMap {
			samples := containerMetric.GetSamples(FormatEnum.MetricTypeCPUUsageSecondsPercentage)
			cpuSampleList = append(cpuSampleList, samples)
		}
	}
	err := containerCpuRepo.CreateMetrics(cpuSampleList)
	if err != nil {
		scope.Error(err.Error())
		return err
	}

	// Write container memory metrics
	containerMemoryRepo := RepoInfluxMetric.NewContainerMemoryRepositoryWithConfig(p.InfluxDBConfig)
	memorySampleList := make([]*DaoMetricTypes.ContainerMetricSample, 0)
	for _, podMetric := range metrics.MetricMap {
		for _, containerMetric := range podMetric.ContainerMetricMap.MetricMap {
			samples := containerMetric.GetSamples(FormatEnum.MetricTypeMemoryUsageBytes)
			memorySampleList = append(memorySampleList, samples)
		}
	}
	err = containerMemoryRepo.CreateMetrics(memorySampleList)
	if err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}

func (p *PodMetrics) ListMetrics(request DaoMetricTypes.ListPodMetricsRequest) (DaoMetricTypes.PodMetricMap, error) {
	podMetricMap := DaoMetricTypes.NewPodMetricMap()

	// Read container cpu metrics
	containerCpuRepo := RepoInfluxMetric.NewContainerCpuRepositoryWithConfig(p.InfluxDBConfig)
	cpuMetrics, err := containerCpuRepo.ListMetrics(request)
	if err != nil {
		scope.Error(err.Error())
		return DaoMetricTypes.NewPodMetricMap(), err
	}
	for _, nodeMetric := range cpuMetrics {
		podMetricMap.AddContainerMetric(nodeMetric)
	}

	// Read node memory metrics
	containerMemoryRepo := RepoInfluxMetric.NewContainerMemoryRepositoryWithConfig(p.InfluxDBConfig)
	memoryMetrics, err := containerMemoryRepo.ListMetrics(request)
	if err != nil {
		scope.Error(err.Error())
		return DaoMetricTypes.NewPodMetricMap(), err
	}
	for _, nodeMetric := range memoryMetrics {
		podMetricMap.AddContainerMetric(nodeMetric)
	}

	return podMetricMap, nil
}
