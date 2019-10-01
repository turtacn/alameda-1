package nvidia

import (
	DatahubMetric "github.com/containers-ai/alameda/datahub/pkg/metric"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
)

type GpuPredictionMap map[DatahubMetric.GpuMetricType][]*GpuPrediction

type GpuPrediction struct {
	Gpu
	Granularity  int64
	ModelId      string
	PredictionId string
	Metrics      []DatahubMetric.Sample
}

type PredictionsDAO interface {
	CreatePredictions(GpuPredictionMap) error
	ListPredictions(host, minorNumber, modelId, predictionId, granularity string, condition *DBCommon.QueryCondition) (GpuPredictionMap, error)
}

func NewGpuPrediction() *GpuPrediction {
	gpu := &GpuPrediction{}
	gpu.Metrics = make([]DatahubMetric.Sample, 0)
	return gpu
}

func NewGpuPredictionMap() GpuPredictionMap {
	return GpuPredictionMap{}
}

func (p *GpuPredictionMap) AddGpuPrediction(gpu *Gpu, granularity int64, modelId, predictionId string, metricType DatahubMetric.GpuMetricType, sample DatahubMetric.Sample) {
	if _, exist := (*p)[metricType]; !exist {
		(*p)[metricType] = make([]*GpuPrediction, 0)
	}

	gpuPrediction := NewGpuPrediction()
	found := false
	for _, gpuPrediction = range (*p)[metricType] {
		if gpuPrediction.Uuid == gpu.Uuid {
			if gpuPrediction.ModelId == modelId && gpuPrediction.PredictionId == predictionId {
				found = true
				break
			}
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
		gpuPrediction.ModelId = modelId
		gpuPrediction.PredictionId = predictionId

		(*p)[metricType] = append((*p)[metricType], gpuPrediction)
	}

	gpuPrediction.Metrics = append(gpuPrediction.Metrics, sample)
}
