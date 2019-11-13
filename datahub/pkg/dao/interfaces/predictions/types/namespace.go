package types

import (
	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
)

type NamespacePredictionsDAO interface {
	CreatePredictions(NamespacePredictionMap) error
	ListPredictions(ListNamespacePredictionsRequest) (NamespacePredictionMap, error)
}

// NamespacePrediction Prediction model to represent one namespace Prediction
type NamespacePrediction struct {
	ObjectMeta           metadata.ObjectMeta
	PredictionRaw        map[enumconv.MetricType]*types.PredictionMetricData
	PredictionUpperBound map[enumconv.MetricType]*types.PredictionMetricData
	PredictionLowerBound map[enumconv.MetricType]*types.PredictionMetricData
}

// NamespacesPredictionMap Namespaces' Prediction map
type NamespacePredictionMap struct {
	MetricMap map[string]*NamespacePrediction
}

// ListNamespacePredictionsRequest ListNamespacePredictionsRequest
type ListNamespacePredictionsRequest struct {
	common.QueryCondition
	ObjectMeta   []metadata.ObjectMeta
	ModelId      string
	PredictionId string
	Granularity  int64
}

func NewNamespacePrediction() *NamespacePrediction {
	namespacePrediction := &NamespacePrediction{}
	namespacePrediction.PredictionRaw = make(map[enumconv.MetricType]*types.PredictionMetricData)
	namespacePrediction.PredictionUpperBound = make(map[enumconv.MetricType]*types.PredictionMetricData)
	namespacePrediction.PredictionLowerBound = make(map[enumconv.MetricType]*types.PredictionMetricData)
	return namespacePrediction
}

func NewNamespacePredictionMap() NamespacePredictionMap {
	namespacePredictionMap := NamespacePredictionMap{}
	namespacePredictionMap.MetricMap = make(map[string]*NamespacePrediction)
	return namespacePredictionMap
}

func NewListNamespacePredictionRequest() ListNamespacePredictionsRequest {
	request := ListNamespacePredictionsRequest{}
	request.ObjectMeta = make([]metadata.ObjectMeta, 0)
	return request
}

func (n *NamespacePrediction) AddRawSample(metricType enumconv.MetricType, granularity int64, sample types.PredictionSample) {
	if _, exist := n.PredictionRaw[metricType]; !exist {
		n.PredictionRaw[metricType] = types.NewPredictionMetricData()
		n.PredictionRaw[metricType].Granularity = granularity
	}
	n.PredictionRaw[metricType].Data = append(n.PredictionRaw[metricType].Data, sample)
}

func (n *NamespacePrediction) AddUpperBoundSample(metricType enumconv.MetricType, granularity int64, sample types.PredictionSample) {
	if _, exist := n.PredictionUpperBound[metricType]; !exist {
		n.PredictionUpperBound[metricType] = types.NewPredictionMetricData()
		n.PredictionUpperBound[metricType].Granularity = granularity
	}
	n.PredictionUpperBound[metricType].Data = append(n.PredictionUpperBound[metricType].Data, sample)
}

func (n *NamespacePrediction) AddLowerBoundSample(metricType enumconv.MetricType, granularity int64, sample types.PredictionSample) {
	if _, exist := n.PredictionLowerBound[metricType]; !exist {
		n.PredictionLowerBound[metricType] = types.NewPredictionMetricData()
		n.PredictionLowerBound[metricType].Granularity = granularity
	}
	n.PredictionLowerBound[metricType].Data = append(n.PredictionLowerBound[metricType].Data, sample)
}

// Merge Merge current NamespacePrediction with input NamespacePrediction
func (n *NamespacePrediction) Merge(in *NamespacePrediction) {
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

// AddNamespacePrediction Add namespace Prediction into namespacesPredictionMap
func (n *NamespacePredictionMap) AddNamespacePrediction(namespacePrediction *NamespacePrediction) {
	namespaceName := namespacePrediction.ObjectMeta.Name
	if existNamespacePrediction, exist := n.MetricMap[namespaceName]; exist {
		existNamespacePrediction.Merge(namespacePrediction)
	} else {
		n.MetricMap[namespaceName] = namespacePrediction
	}
}
