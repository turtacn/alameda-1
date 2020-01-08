package nvidia

import (
	DaoGpu "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/gpu/influxdb"
	RepoInfluxGpuMetric "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/gpu/nvidia/metrics"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	Utils "github.com/containers-ai/alameda/datahub/pkg/utils"
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

func (p Metric) ListMetrics(host, minorNumber string, metricTypes []FormatEnum.MetricType, condition *DBCommon.QueryCondition) (DaoGpu.GpuMetricMap, error) {
	gpuMetricMap := DaoGpu.NewGpuMetricMap()

	// Pack duty cycle metrics
	if Utils.SliceContains(metricTypes, FormatEnum.MetricTypeDutyCycle) {
		if DaoGpu.GpuMetricUsedMap[FormatEnum.TypeGpuDutyCycle] {
			dutyCycleRepo := RepoInfluxGpuMetric.NewDutyCycleRepositoryWithConfig(p.InfluxDBConfig)
			dutyCycleMetrics, err := dutyCycleRepo.ListMetrics(host, minorNumber, condition)
			if err != nil {
				return DaoGpu.NewGpuMetricMap(), err
			}
			for _, metrics := range dutyCycleMetrics {
				sample := FormatTypes.Sample{Timestamp: metrics.Time, Value: strconv.FormatFloat(*metrics.Value, 'f', -1, 64)}
				gpu := DaoGpu.NewGpu()
				gpu.Name = *metrics.Name
				gpu.Uuid = *metrics.Uuid
				gpu.Metadata.Host = *metrics.Host
				gpu.Metadata.Instance = *metrics.Instance
				gpu.Metadata.Job = *metrics.Job
				gpu.Metadata.MinorNumber = *metrics.MinorNumber

				gpuMetricMap.AddGpuMetric(gpu, FormatEnum.TypeGpuDutyCycle, sample)
			}
		}
	}

	// Pack memory used bytes metrics
	if Utils.SliceContains(metricTypes, FormatEnum.MetricTypeMemoryUsageBytes) {
		if DaoGpu.GpuMetricUsedMap[FormatEnum.TypeGpuMemoryUsedBytes] {
			memoryUsedRepo := RepoInfluxGpuMetric.NewMemoryUsedBytesRepositoryWithConfig(p.InfluxDBConfig)
			memoryUsedMetrics, err := memoryUsedRepo.ListMetrics(host, minorNumber, condition)
			if err != nil {
				return DaoGpu.NewGpuMetricMap(), err
			}
			for _, metrics := range memoryUsedMetrics {
				sample := FormatTypes.Sample{Timestamp: metrics.Time, Value: strconv.FormatFloat(*metrics.Value, 'f', -1, 64)}
				gpu := DaoGpu.NewGpu()
				gpu.Name = *metrics.Name
				gpu.Uuid = *metrics.Uuid
				gpu.Metadata.Host = *metrics.Host
				gpu.Metadata.Instance = *metrics.Instance
				gpu.Metadata.Job = *metrics.Job
				gpu.Metadata.MinorNumber = *metrics.MinorNumber

				gpuMetricMap.AddGpuMetric(gpu, FormatEnum.TypeGpuMemoryUsedBytes, sample)
			}
		}
	}

	// Pack power usage milli watts metrics
	if Utils.SliceContains(metricTypes, FormatEnum.MetricTypePowerUsageWatts) {
		if DaoGpu.GpuMetricUsedMap[FormatEnum.TypeGpuPowerUsageMilliWatts] {
			powerUsageRepo := RepoInfluxGpuMetric.NewPowerUsageMilliWattsRepositoryWithConfig(p.InfluxDBConfig)
			powerUsageMetrics, err := powerUsageRepo.ListMetrics(host, minorNumber, condition)
			if err != nil {
				return DaoGpu.NewGpuMetricMap(), err
			}
			for _, metrics := range powerUsageMetrics {
				sample := FormatTypes.Sample{Timestamp: metrics.Time, Value: strconv.FormatFloat(*metrics.Value, 'f', -1, 64)}
				gpu := DaoGpu.NewGpu()
				gpu.Name = *metrics.Name
				gpu.Uuid = *metrics.Uuid
				gpu.Metadata.Host = *metrics.Host
				gpu.Metadata.Instance = *metrics.Instance
				gpu.Metadata.Job = *metrics.Job
				gpu.Metadata.MinorNumber = *metrics.MinorNumber

				gpuMetricMap.AddGpuMetric(gpu, FormatEnum.TypeGpuPowerUsageMilliWatts, sample)
			}
		}
	}

	// Pack temperature celsius metrics
	if Utils.SliceContains(metricTypes, FormatEnum.MetricTypeTemperatureCelsius) {
		if DaoGpu.GpuMetricUsedMap[FormatEnum.TypeGpuTemperatureCelsius] {
			temperatureRepo := RepoInfluxGpuMetric.NewTemperatureCelsiusRepositoryWithConfig(p.InfluxDBConfig)
			temperatureMetrics, err := temperatureRepo.ListMetrics(host, minorNumber, condition)
			if err != nil {
				return DaoGpu.NewGpuMetricMap(), err
			}
			for _, metrics := range temperatureMetrics {
				sample := FormatTypes.Sample{Timestamp: metrics.Time, Value: strconv.FormatFloat(*metrics.Value, 'f', -1, 64)}
				gpu := DaoGpu.NewGpu()
				gpu.Name = *metrics.Name
				gpu.Uuid = *metrics.Uuid
				gpu.Metadata.Host = *metrics.Host
				gpu.Metadata.Instance = *metrics.Instance
				gpu.Metadata.Job = *metrics.Job
				gpu.Metadata.MinorNumber = *metrics.MinorNumber

				gpuMetricMap.AddGpuMetric(gpu, FormatEnum.TypeGpuTemperatureCelsius, sample)
			}
		}
	}

	return gpuMetricMap, nil
}
