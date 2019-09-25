package dispatcher

import (
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

type modelInfo struct {
	podModel  `json:",inline"`
	nodeModel `json:",inline"`
	gpuModel  `json:",inline"`

	ModelMetrics    []datahub_v1alpha1.MetricType `json:"modelMetrics,omitempty"`
	Timestamp       int64                         `json:"timestamp"`
	CreateTimestamp int64                         `json:"createTimestamp"`
}

func (modelInfo *modelInfo) SetTimeStamp(ts int64) {
	modelInfo.Timestamp = ts
}

func (modelInfo *modelInfo) GetTimeStamp() int64 {
	return modelInfo.Timestamp
}

func (modelInfo *modelInfo) SetCreateTimeStamp(ts int64) {
	modelInfo.CreateTimestamp = ts
}

func (modelInfo *modelInfo) GetCreateTimeStamp() int64 {
	return modelInfo.CreateTimestamp
}
