package influxdb

import (
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
)

type GpuPredictionMap map[FormatEnum.GpuMetricType][]*GpuPrediction

type GpuPrediction struct {
	Gpu
	Granularity int64
	Metrics     []FormatTypes.PredictionSample
}

type PredictionsDAO interface {
	CreatePredictions(GpuPredictionMap) error
	ListPredictions(host, minorNumber, modelId, predictionId, granularity string, condition *DBCommon.QueryCondition) (GpuPredictionMap, error)
}

func NewGpuPrediction() *GpuPrediction {
	gpu := &GpuPrediction{}
	gpu.Metrics = make([]FormatTypes.PredictionSample, 0)
	return gpu
}

func NewGpuPredictionMap() GpuPredictionMap {
	return GpuPredictionMap{}
}

func (p *GpuPredictionMap) AddGpuPrediction(gpu *Gpu, granularity int64, metricType FormatEnum.GpuMetricType, sample FormatTypes.PredictionSample) {
	if _, exist := (*p)[metricType]; !exist {
		(*p)[metricType] = make([]*GpuPrediction, 0)
	}

	gpuPrediction := NewGpuPrediction()
	found := false
	for _, gpuPrediction = range (*p)[metricType] {
		if gpuPrediction.Uuid == gpu.Uuid {
			found = true
			break
		}
	}

	if found == false {
		gpuPrediction = NewGpuPrediction()
		gpuPrediction.Name = gpu.Name
		gpuPrediction.Uuid = gpu.Uuid
		gpuPrediction.Metadata.Host = gpu.Metadata.Host
		gpuPrediction.Metadata.Instance = gpu.Metadata.Instance
		gpuPrediction.Metadata.Job = gpu.Metadata.Job
		gpuPrediction.Metadata.MinorNumber = gpu.Metadata.MinorNumber
		gpuPrediction.Granularity = granularity

		(*p)[metricType] = append((*p)[metricType], gpuPrediction)
	}

	gpuPrediction.Metrics = append(gpuPrediction.Metrics, sample)
}
