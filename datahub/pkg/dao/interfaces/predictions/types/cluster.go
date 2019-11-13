package types

import (
	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
)

type ClusterPredictionsDAO interface {
	CreatePredictions(ClusterPredictionMap) error
	ListPredictions(ListClusterPredictionsRequest) (ClusterPredictionMap, error)
}

// ClusterPrediction Prediction model to represent one cluster Prediction
type ClusterPrediction struct {
	ObjectMeta           metadata.ObjectMeta
	PredictionRaw        map[enumconv.MetricType]*types.PredictionMetricData
	PredictionUpperBound map[enumconv.MetricType]*types.PredictionMetricData
	PredictionLowerBound map[enumconv.MetricType]*types.PredictionMetricData
}

// ClustersPredictionMap Clusters' Prediction map
type ClusterPredictionMap struct {
	MetricMap map[string]*ClusterPrediction
}

// ListClusterPredictionsRequest ListClusterPredictionsRequest
type ListClusterPredictionsRequest struct {
	common.QueryCondition
	ObjectMeta   []metadata.ObjectMeta
	ModelId      string
	PredictionId string
	Granularity  int64
}

func NewClusterPrediction() *ClusterPrediction {
	clusterPrediction := &ClusterPrediction{}
	clusterPrediction.PredictionRaw = make(map[enumconv.MetricType]*types.PredictionMetricData)
	clusterPrediction.PredictionUpperBound = make(map[enumconv.MetricType]*types.PredictionMetricData)
	clusterPrediction.PredictionLowerBound = make(map[enumconv.MetricType]*types.PredictionMetricData)
	return clusterPrediction
}

func NewClusterPredictionMap() ClusterPredictionMap {
	clusterPredictionMap := ClusterPredictionMap{}
	clusterPredictionMap.MetricMap = make(map[string]*ClusterPrediction)
	return clusterPredictionMap
}

func NewListClusterPredictionRequest() ListClusterPredictionsRequest {
	request := ListClusterPredictionsRequest{}
	request.ObjectMeta = make([]metadata.ObjectMeta, 0)
	return request
}

func (n *ClusterPrediction) AddRawSample(metricType enumconv.MetricType, granularity int64, sample types.PredictionSample) {
	if _, exist := n.PredictionRaw[metricType]; !exist {
		n.PredictionRaw[metricType] = types.NewPredictionMetricData()
		n.PredictionRaw[metricType].Granularity = granularity
	}
	n.PredictionRaw[metricType].Data = append(n.PredictionRaw[metricType].Data, sample)
}

func (n *ClusterPrediction) AddUpperBoundSample(metricType enumconv.MetricType, granularity int64, sample types.PredictionSample) {
	if _, exist := n.PredictionUpperBound[metricType]; !exist {
		n.PredictionUpperBound[metricType] = types.NewPredictionMetricData()
		n.PredictionUpperBound[metricType].Granularity = granularity
	}
	n.PredictionUpperBound[metricType].Data = append(n.PredictionUpperBound[metricType].Data, sample)
}

func (n *ClusterPrediction) AddLowerBoundSample(metricType enumconv.MetricType, granularity int64, sample types.PredictionSample) {
	if _, exist := n.PredictionLowerBound[metricType]; !exist {
		n.PredictionLowerBound[metricType] = types.NewPredictionMetricData()
		n.PredictionLowerBound[metricType].Granularity = granularity
	}
	n.PredictionLowerBound[metricType].Data = append(n.PredictionLowerBound[metricType].Data, sample)
}

// Merge Merge current ClusterPrediction with input ClusterPrediction
func (n *ClusterPrediction) Merge(in *ClusterPrediction) {
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

// AddClusterPrediction Add cluster Prediction into clustersPredictionMap
func (n *ClusterPredictionMap) AddClusterPrediction(clusterPrediction *ClusterPrediction) {
	clusterName := clusterPrediction.ObjectMeta.Name
	if existClusterPrediction, exist := n.MetricMap[clusterName]; exist {
		existClusterPrediction.Merge(clusterPrediction)
	} else {
		n.MetricMap[clusterName] = clusterPrediction
	}
}
