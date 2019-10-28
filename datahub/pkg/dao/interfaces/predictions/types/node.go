package types

import (
	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
)

type NodePredictionsDAO interface {
	CreatePredictions(NodePredictionMap) error
	ListPredictions(ListNodePredictionsRequest) (NodePredictionMap, error)
	FillPredictions(predictions []*ApiPredictions.NodePrediction, fillDays int64) error
}

// NodePrediction Prediction model to represent one node Prediction
type NodePrediction struct {
	NodeName             metadata.NodeName
	IsScheduled          bool
	PredictionRaw        map[enumconv.MetricType]*types.PredictionMetricData
	PredictionUpperBound map[enumconv.MetricType]*types.PredictionMetricData
	PredictionLowerBound map[enumconv.MetricType]*types.PredictionMetricData
}

// NodesPredictionMap Nodes' Prediction map
type NodePredictionMap struct {
	MetricMap map[metadata.NodeName]*NodePrediction
}

// ListNodePredictionsRequest ListNodePredictionsRequest
type ListNodePredictionsRequest struct {
	common.QueryCondition
	NodeNames    []string
	ModelId      string
	PredictionId string
	Granularity  int64
}

func NewNodePrediction() *NodePrediction {
	nodePrediction := &NodePrediction{}
	nodePrediction.PredictionRaw = make(map[enumconv.MetricType]*types.PredictionMetricData)
	nodePrediction.PredictionUpperBound = make(map[enumconv.MetricType]*types.PredictionMetricData)
	nodePrediction.PredictionLowerBound = make(map[enumconv.MetricType]*types.PredictionMetricData)
	return nodePrediction
}

func NewNodePredictionMap() NodePredictionMap {
	nodePredictionMap := NodePredictionMap{}
	nodePredictionMap.MetricMap = make(map[metadata.NodeName]*NodePrediction)
	return nodePredictionMap
}

func (n *NodePrediction) AddRawSample(metricType enumconv.MetricType, granularity int64, sample types.PredictionSample) {
	if _, exist := n.PredictionRaw[metricType]; !exist {
		n.PredictionRaw[metricType] = types.NewPredictionMetricData()
		n.PredictionRaw[metricType].Granularity = granularity
	}
	n.PredictionRaw[metricType].Data = append(n.PredictionRaw[metricType].Data, sample)
}

func (n *NodePrediction) AddUpperBoundSample(metricType enumconv.MetricType, granularity int64, sample types.PredictionSample) {
	if _, exist := n.PredictionUpperBound[metricType]; !exist {
		n.PredictionUpperBound[metricType] = types.NewPredictionMetricData()
		n.PredictionUpperBound[metricType].Granularity = granularity
	}
	n.PredictionUpperBound[metricType].Data = append(n.PredictionUpperBound[metricType].Data, sample)
}

func (n *NodePrediction) AddLowerBoundSample(metricType enumconv.MetricType, granularity int64, sample types.PredictionSample) {
	if _, exist := n.PredictionLowerBound[metricType]; !exist {
		n.PredictionLowerBound[metricType] = types.NewPredictionMetricData()
		n.PredictionLowerBound[metricType].Granularity = granularity
	}
	n.PredictionLowerBound[metricType].Data = append(n.PredictionLowerBound[metricType].Data, sample)
}

// Merge Merge current NodePrediction with input NodePrediction
func (n *NodePrediction) Merge(in *NodePrediction) {
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

// AddNodePrediction Add node Prediction into NodesPredictionMap
func (n *NodePredictionMap) AddNodePrediction(nodePrediction *NodePrediction) {
	nodeName := nodePrediction.NodeName
	if existNodePrediction, exist := n.MetricMap[nodeName]; exist {
		existNodePrediction.Merge(nodePrediction)
	} else {
		n.MetricMap[nodeName] = nodePrediction
	}
}
