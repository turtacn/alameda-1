package impl

import (
	DaoGpu "github.com/containers-ai/alameda/datahub/pkg/dao/gpu/nvidia"
	DatahubMetric "github.com/containers-ai/alameda/datahub/pkg/metric"
	RepoInfluxGpuMetric "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb/gpu/nvidia/metric"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"strconv"
)

type Metric struct {
	InfluxDBConfig InternalInflux.Config
}

func NewMetricWithConfig(config InternalInflux.Config) DaoGpu.MetricsDAO {
	return Metric{InfluxDBConfig: config}
}

func (p Metric) ListMetrics(host, minorNumber string, condition *DBCommon.QueryCondition) (DaoGpu.GpuMetricMap, error) {
	gpuMetricMap := DaoGpu.NewGpuMetricMap()

	// Pack duty cycle metrics
	dutyCycleRepo := RepoInfluxGpuMetric.NewDutyCycleRepositoryWithConfig(p.InfluxDBConfig)
	dutyCycleMetrics, err := dutyCycleRepo.ListMetrics(host, minorNumber, condition)
	if err != nil {
		return DaoGpu.NewGpuMetricMap(), err
	}
	for _, metrics := range dutyCycleMetrics {
		sample := DatahubMetric.Sample{Timestamp: metrics.Time, Value: strconv.FormatFloat(*metrics.Value, 'f', -1, 64)}
		gpu := DaoGpu.NewGpu()
		gpu.Name = *metrics.Name
		gpu.Uuid = *metrics.Uuid
		gpu.Metadata.Host = *metrics.Host
		gpu.Metadata.Instance = *metrics.Instance
		gpu.Metadata.Job = *metrics.Job
		gpu.Metadata.MinorNumber = *metrics.MinorNumber

		gpuMetricMap.AddGpuMetric(gpu, DatahubMetric.TypeGpuDutyCycle, sample)
	}

	// Pack memory used bytes metrics
	memoryUsedRepo := RepoInfluxGpuMetric.NewMemoryUsedBytesRepositoryWithConfig(p.InfluxDBConfig)
	memoryUsedMetrics, err := memoryUsedRepo.ListMetrics(host, minorNumber, condition)
	if err != nil {
		return DaoGpu.NewGpuMetricMap(), err
	}
	for _, metrics := range memoryUsedMetrics {
		sample := DatahubMetric.Sample{Timestamp: metrics.Time, Value: strconv.FormatFloat(*metrics.Value, 'f', -1, 64)}
		gpu := DaoGpu.NewGpu()
		gpu.Name = *metrics.Name
		gpu.Uuid = *metrics.Uuid
		gpu.Metadata.Host = *metrics.Host
		gpu.Metadata.Instance = *metrics.Instance
		gpu.Metadata.Job = *metrics.Job
		gpu.Metadata.MinorNumber = *metrics.MinorNumber

		gpuMetricMap.AddGpuMetric(gpu, DatahubMetric.TypeGpuMemoryUsedBytes, sample)
	}

	// Pack power usage milli watts metrics
	powerUsageRepo := RepoInfluxGpuMetric.NewPowerUsageMilliWattsRepositoryWithConfig(p.InfluxDBConfig)
	powerUsageMetrics, err := powerUsageRepo.ListMetrics(host, minorNumber, condition)
	if err != nil {
		return DaoGpu.NewGpuMetricMap(), err
	}
	for _, metrics := range powerUsageMetrics {
		sample := DatahubMetric.Sample{Timestamp: metrics.Time, Value: strconv.FormatFloat(*metrics.Value, 'f', -1, 64)}
		gpu := DaoGpu.NewGpu()
		gpu.Name = *metrics.Name
		gpu.Uuid = *metrics.Uuid
		gpu.Metadata.Host = *metrics.Host
		gpu.Metadata.Instance = *metrics.Instance
		gpu.Metadata.Job = *metrics.Job
		gpu.Metadata.MinorNumber = *metrics.MinorNumber

		gpuMetricMap.AddGpuMetric(gpu, DatahubMetric.TypeGpuPowerUsageMilliWatts, sample)
	}

	// Pack temperature celsius metrics
	temperatureRepo := RepoInfluxGpuMetric.NewTemperatureCelsiusRepositoryWithConfig(p.InfluxDBConfig)
	temperatureMetrics, err := temperatureRepo.ListMetrics(host, minorNumber, condition)
	if err != nil {
		return DaoGpu.NewGpuMetricMap(), err
	}
	for _, metrics := range temperatureMetrics {
		sample := DatahubMetric.Sample{Timestamp: metrics.Time, Value: strconv.FormatFloat(*metrics.Value, 'f', -1, 64)}
		gpu := DaoGpu.NewGpu()
		gpu.Name = *metrics.Name
		gpu.Uuid = *metrics.Uuid
		gpu.Metadata.Host = *metrics.Host
		gpu.Metadata.Instance = *metrics.Instance
		gpu.Metadata.Job = *metrics.Job
		gpu.Metadata.MinorNumber = *metrics.MinorNumber

		gpuMetricMap.AddGpuMetric(gpu, DatahubMetric.TypeGpuTemperatureCelsius, sample)
	}

	return gpuMetricMap, nil
}
