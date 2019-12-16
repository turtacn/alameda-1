package types

import (
	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
	"strings"
)

type ApplicationPredictionsDAO interface {
	CreatePredictions(ApplicationPredictionMap) error
	ListPredictions(ListApplicationPredictionsRequest) (ApplicationPredictionMap, error)
}

// ApplicationPrediction Prediction model to represent one application Prediction
type ApplicationPrediction struct {
	ObjectMeta           metadata.ObjectMeta
	PredictionRaw        map[enumconv.MetricType]*types.PredictionMetricData
	PredictionUpperBound map[enumconv.MetricType]*types.PredictionMetricData
	PredictionLowerBound map[enumconv.MetricType]*types.PredictionMetricData
}

// ApplicationsPredictionMap Applications' Prediction map
type ApplicationPredictionMap struct {
	MetricMap map[string]*ApplicationPrediction
}

// ListApplicationPredictionsRequest ListApplicationPredictionsRequest
type ListApplicationPredictionsRequest struct {
	common.QueryCondition
	ObjectMeta   []metadata.ObjectMeta
	ModelId      string
	PredictionId string
	Granularity  int64
}

func NewApplicationPrediction() *ApplicationPrediction {
	applicationPrediction := &ApplicationPrediction{}
	applicationPrediction.PredictionRaw = make(map[enumconv.MetricType]*types.PredictionMetricData)
	applicationPrediction.PredictionUpperBound = make(map[enumconv.MetricType]*types.PredictionMetricData)
	applicationPrediction.PredictionLowerBound = make(map[enumconv.MetricType]*types.PredictionMetricData)
	return applicationPrediction
}

func NewApplicationPredictionMap() ApplicationPredictionMap {
	applicationPredictionMap := ApplicationPredictionMap{}
	applicationPredictionMap.MetricMap = make(map[string]*ApplicationPrediction)
	return applicationPredictionMap
}

func NewListApplicationPredictionRequest() ListApplicationPredictionsRequest {
	request := ListApplicationPredictionsRequest{}
	request.ObjectMeta = make([]metadata.ObjectMeta, 0)
	return request
}

func (n *ApplicationPrediction) Identifier() string {
	if !n.ObjectMeta.IsEmpty() {
		valueList := make([]string, 0)
		valueList = append(valueList, n.ObjectMeta.ClusterName)
		valueList = append(valueList, n.ObjectMeta.Namespace)
		valueList = append(valueList, n.ObjectMeta.Name)
		return strings.Join(valueList, "/")
	}
	return ""
}

func (n *ApplicationPrediction) AddRawSample(metricType enumconv.MetricType, granularity int64, sample types.PredictionSample) {
	if _, exist := n.PredictionRaw[metricType]; !exist {
		n.PredictionRaw[metricType] = types.NewPredictionMetricData()
		n.PredictionRaw[metricType].Granularity = granularity
	}
	n.PredictionRaw[metricType].Data = append(n.PredictionRaw[metricType].Data, sample)
}

func (n *ApplicationPrediction) AddUpperBoundSample(metricType enumconv.MetricType, granularity int64, sample types.PredictionSample) {
	if _, exist := n.PredictionUpperBound[metricType]; !exist {
		n.PredictionUpperBound[metricType] = types.NewPredictionMetricData()
		n.PredictionUpperBound[metricType].Granularity = granularity
	}
	n.PredictionUpperBound[metricType].Data = append(n.PredictionUpperBound[metricType].Data, sample)
}

func (n *ApplicationPrediction) AddLowerBoundSample(metricType enumconv.MetricType, granularity int64, sample types.PredictionSample) {
	if _, exist := n.PredictionLowerBound[metricType]; !exist {
		n.PredictionLowerBound[metricType] = types.NewPredictionMetricData()
		n.PredictionLowerBound[metricType].Granularity = granularity
	}
	n.PredictionLowerBound[metricType].Data = append(n.PredictionLowerBound[metricType].Data, sample)
}

// Merge Merge current ApplicationPrediction with input ApplicationPrediction
func (n *ApplicationPrediction) Merge(in *ApplicationPrediction) {
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

// AddApplicationPrediction Add application Prediction into ApplicationsPredictionMap
func (n *ApplicationPredictionMap) AddApplicationPrediction(applicationPrediction *ApplicationPrediction) {
	identifier := applicationPrediction.Identifier()
	if existApplicationPrediction, exist := n.MetricMap[identifier]; exist {
		existApplicationPrediction.Merge(applicationPrediction)
	} else {
		n.MetricMap[identifier] = applicationPrediction
	}
}
