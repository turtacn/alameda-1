package requests

import (
	DaoPredictionTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	Metadata "github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
	"github.com/golang/protobuf/ptypes"
)

type CreateNodePredictionsRequestExtended struct {
	ApiPredictions.CreateNodePredictionsRequest
}

func (r *CreateNodePredictionsRequestExtended) Validate() error {
	return nil
}

func (r *CreateNodePredictionsRequestExtended) ProducePredictions() DaoPredictionTypes.NodePredictionMap {
	nodePredictionMap := DaoPredictionTypes.NewNodePredictionMap()

	for _, node := range r.GetNodePredictions() {
		// Normalize request
		objectMeta := NewObjectMeta(node.GetObjectMeta())
		objectMeta.Namespace = ""
		objectMeta.NodeName = ""

		nodePrediction := DaoPredictionTypes.NewNodePrediction()
		nodePrediction.ObjectMeta = objectMeta
		nodePrediction.IsScheduled = node.GetIsScheduled()

		// Handle predicted raw data
		for _, data := range node.GetPredictedRawData() {
			metricType := MetricTypeNameMap[data.GetMetricType()]
			granularity := data.GetGranularity()
			for _, sample := range data.GetData() {
				timestamp, err := ptypes.Timestamp(sample.GetTime())
				if err != nil {
					scope.Error(" failed: " + err.Error())
				}
				sample := FormatTypes.PredictionSample{
					Timestamp:    timestamp,
					Value:        sample.GetNumValue(),
					ModelId:      sample.GetModelId(),
					PredictionId: sample.GetPredictionId(),
				}
				nodePrediction.AddRawSample(metricType, granularity, sample)
			}
		}

		// Handle predicted upper bound data
		for _, data := range node.GetPredictedUpperboundData() {
			metricType := MetricTypeNameMap[data.GetMetricType()]
			granularity := data.GetGranularity()
			for _, sample := range data.GetData() {
				timestamp, err := ptypes.Timestamp(sample.GetTime())
				if err != nil {
					scope.Error(" failed: " + err.Error())
				}
				sample := FormatTypes.PredictionSample{
					Timestamp:    timestamp,
					Value:        sample.GetNumValue(),
					ModelId:      sample.GetModelId(),
					PredictionId: sample.GetPredictionId(),
				}
				nodePrediction.AddUpperBoundSample(metricType, granularity, sample)
			}
		}

		// Handle predicted lower bound data
		for _, data := range node.GetPredictedLowerboundData() {
			metricType := MetricTypeNameMap[data.GetMetricType()]
			granularity := data.GetGranularity()
			for _, sample := range data.GetData() {
				timestamp, err := ptypes.Timestamp(sample.GetTime())
				if err != nil {
					scope.Error(" failed: " + err.Error())
				}
				sample := FormatTypes.PredictionSample{
					Timestamp:    timestamp,
					Value:        sample.GetNumValue(),
					ModelId:      sample.GetModelId(),
					PredictionId: sample.GetPredictionId(),
				}
				nodePrediction.AddLowerBoundSample(metricType, granularity, sample)
			}
		}

		nodePredictionMap.AddNodePrediction(nodePrediction)
	}

	return nodePredictionMap
}

type ListNodePredictionsRequestExtended struct {
	Request *ApiPredictions.ListNodePredictionsRequest
}

func (r *ListNodePredictionsRequestExtended) Validate() error {
	return nil
}

func (r *ListNodePredictionsRequestExtended) ProduceRequest() DaoPredictionTypes.ListNodePredictionsRequest {
	request := DaoPredictionTypes.NewListNodePredictionRequest()
	request.QueryCondition = QueryConditionExtend{r.Request.GetQueryCondition()}.QueryCondition()
	request.Granularity = 30
	request.ModelId = r.Request.GetModelId()
	request.PredictionId = r.Request.GetPredictionId()
	if r.Request.GetGranularity() != 0 {
		request.Granularity = r.Request.GetGranularity()
	}
	if r.Request.GetObjectMeta() != nil {
		for _, meta := range r.Request.GetObjectMeta() {
			// Normalize request
			objectMeta := NewObjectMeta(meta)
			objectMeta.Namespace = ""
			objectMeta.NodeName = ""

			if objectMeta.IsEmpty() {
				request.ObjectMeta = make([]Metadata.ObjectMeta, 0)
				return request
			}
			request.ObjectMeta = append(request.ObjectMeta, objectMeta)
		}
	}
	return request
}
