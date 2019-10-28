package influxdb

import (
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
)

type GpuMetricMap map[string]*GpuMetric

type GpuMetric struct {
	Gpu
	Metrics map[FormatEnum.GpuMetricType][]FormatTypes.Sample
}

type MetricsDAO interface {
	ListMetrics(host, minorNumber string, condition *DBCommon.QueryCondition) (GpuMetricMap, error)
}

func NewGpuMetric() *GpuMetric {
	gpuMetric := &GpuMetric{}
	gpuMetric.Metrics = make(map[FormatEnum.GpuMetricType][]FormatTypes.Sample)
	return gpuMetric
}

func NewGpuMetricMap() GpuMetricMap {
	return GpuMetricMap{}
}

func (p *GpuMetricMap) AddGpuMetric(gpu *Gpu, metricType FormatEnum.GpuMetricType, sample FormatTypes.Sample) {
	if _, exist := (*p)[gpu.Uuid]; !exist {
		gpuMetric := NewGpuMetric()
		gpuMetric.Name = gpu.Name
		gpuMetric.Uuid = gpu.Uuid
		gpuMetric.Metadata.Host = gpu.Metadata.Host
		gpuMetric.Metadata.Instance = gpu.Metadata.Instance
		gpuMetric.Metadata.Job = gpu.Metadata.Job
		gpuMetric.Metadata.MinorNumber = gpu.Metadata.MinorNumber

		(*p)[gpu.Uuid] = gpuMetric
	}

	if _, exist := (*p)[gpu.Uuid].Metrics[metricType]; exist {
		(*p)[gpu.Uuid].Metrics[metricType] = append((*p)[gpu.Uuid].Metrics[metricType], sample)
	} else {
		(*p)[gpu.Uuid].Metrics[metricType] = make([]FormatTypes.Sample, 0)
		(*p)[gpu.Uuid].Metrics[metricType] = append((*p)[gpu.Uuid].Metrics[metricType], sample)
	}
}
