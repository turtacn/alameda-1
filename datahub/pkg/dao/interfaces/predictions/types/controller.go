package types

import (
	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
)

type ControllerPredictionsDAO interface {
	CreatePredictions(ControllerPredictionMap) error
	ListPredictions(ListControllerPredictionsRequest) (ControllerPredictionMap, error)
}

// ControllerPrediction Prediction model to represent one controller Prediction
type ControllerPrediction struct {
	ObjectMeta           metadata.ObjectMeta
	Kind                 string
	PredictionRaw        map[enumconv.MetricType]*types.PredictionMetricData
	PredictionUpperBound map[enumconv.MetricType]*types.PredictionMetricData
	PredictionLowerBound map[enumconv.MetricType]*types.PredictionMetricData
}

// ControllersPredictionMap Controllers' Prediction map
type ControllerPredictionMap struct {
	MetricMap map[string]*ControllerPrediction
}

// ListControllerPredictionsRequest ListControllerPredictionsRequest
type ListControllerPredictionsRequest struct {
	common.QueryCondition
	ObjectMeta   []metadata.ObjectMeta
	ModelId      string
	PredictionId string
	Granularity  int64
	Kind         string
}

func NewControllerPrediction() *ControllerPrediction {
	controllerPrediction := &ControllerPrediction{}
	controllerPrediction.PredictionRaw = make(map[enumconv.MetricType]*types.PredictionMetricData)
	controllerPrediction.PredictionUpperBound = make(map[enumconv.MetricType]*types.PredictionMetricData)
	controllerPrediction.PredictionLowerBound = make(map[enumconv.MetricType]*types.PredictionMetricData)
	return controllerPrediction
}

func NewControllerPredictionMap() ControllerPredictionMap {
	controllerPredictionMap := ControllerPredictionMap{}
	controllerPredictionMap.MetricMap = make(map[string]*ControllerPrediction)
	return controllerPredictionMap
}

func NewListControllerPredictionRequest() ListControllerPredictionsRequest {
	request := ListControllerPredictionsRequest{}
	request.ObjectMeta = make([]metadata.ObjectMeta, 0)
	return request
}

func (n *ControllerPrediction) AddRawSample(metricType enumconv.MetricType, granularity int64, sample types.PredictionSample) {
	if _, exist := n.PredictionRaw[metricType]; !exist {
		n.PredictionRaw[metricType] = types.NewPredictionMetricData()
		n.PredictionRaw[metricType].Granularity = granularity
	}
	n.PredictionRaw[metricType].Data = append(n.PredictionRaw[metricType].Data, sample)
}

func (n *ControllerPrediction) AddUpperBoundSample(metricType enumconv.MetricType, granularity int64, sample types.PredictionSample) {
	if _, exist := n.PredictionUpperBound[metricType]; !exist {
		n.PredictionUpperBound[metricType] = types.NewPredictionMetricData()
		n.PredictionUpperBound[metricType].Granularity = granularity
	}
	n.PredictionUpperBound[metricType].Data = append(n.PredictionUpperBound[metricType].Data, sample)
}

func (n *ControllerPrediction) AddLowerBoundSample(metricType enumconv.MetricType, granularity int64, sample types.PredictionSample) {
	if _, exist := n.PredictionLowerBound[metricType]; !exist {
		n.PredictionLowerBound[metricType] = types.NewPredictionMetricData()
		n.PredictionLowerBound[metricType].Granularity = granularity
	}
	n.PredictionLowerBound[metricType].Data = append(n.PredictionLowerBound[metricType].Data, sample)
}

// Merge Merge current ControllerPrediction with input ControllerPrediction
func (n *ControllerPrediction) Merge(in *ControllerPrediction) {
	for metricType, metrics := range in.PredictionRaw {
		n.PredictionRaw[metricType].Data = append(n.PredictionRaw[metricType].Data, metrics.Data...)
	}

	for metricType, metrics := range in.PredictionUpperBound {
		n.PredictionUpperBound[metricType].Data = append(n.PredictionUpperBound[metricType].Data, metrics.Data...)
	}

	for metricType, metrics := range in.PredictionLowerBound {
		n.PredictionLowerBound[metricType].Data = append(n.PredictionLowerBound[metricType].Data, metrics.Data...)
	}
}

// AddControllerPrediction Add Controller Prediction into ControllersPredictionMap
func (n *ControllerPredictionMap) AddControllerPrediction(controllerPrediction *ControllerPrediction) {
	controllerName := controllerPrediction.ObjectMeta.Name
	if existControllerPrediction, exist := n.MetricMap[controllerName]; exist {
		existControllerPrediction.Merge(controllerPrediction)
	} else {
		n.MetricMap[controllerName] = controllerPrediction
	}
}
