package nvidia

import (
	DatahubMetric "github.com/containers-ai/alameda/datahub/pkg/metric"
)

type PredictionsDAO interface {
	CreatePredictions(GpuPredictionMap) error
}

type GpuPrediction struct {
	Gpu
	Granularity int64
	Metrics     []DatahubMetric.Sample
}

type GpuPredictionMap map[DatahubMetric.GpuMetricType][]*GpuPrediction
