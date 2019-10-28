package types

import (
	"time"
)

type PredictionSample struct {
	Timestamp    time.Time
	Value        string
	ModelId      string
	PredictionId string
}

type PredictionMetricData struct {
	Granularity int64
	Data        []PredictionSample
}

func NewPredictionMetricData() *PredictionMetricData {
	metricData := PredictionMetricData{}
	metricData.Data = make([]PredictionSample, 0)
	return &metricData
}
