package requests

import (
	DaoPredictionTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	Metadata "github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
	"github.com/golang/protobuf/ptypes"
)

type CreateNamespacePredictionsRequestExtended struct {
	ApiPredictions.CreateNamespacePredictionsRequest
}

func (r *CreateNamespacePredictionsRequestExtended) Validate() error {
	return nil
}

func (r *CreateNamespacePredictionsRequestExtended) ProducePredictions() DaoPredictionTypes.NamespacePredictionMap {
	namespacePredictionMap := DaoPredictionTypes.NewNamespacePredictionMap()

	for _, namespace := range r.GetNamespacePredictions() {
		// Normalize request
		objectMeta := NewObjectMeta(namespace.GetObjectMeta())
		objectMeta.Namespace = ""
		objectMeta.NodeName = ""

		namespacePrediction := DaoPredictionTypes.NewNamespacePrediction()
		namespacePrediction.ObjectMeta = objectMeta

		// Handle predicted raw data
		for _, data := range namespace.GetPredictedRawData() {
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
				namespacePrediction.AddRawSample(metricType, granularity, sample)
			}
		}

		// Handle predicted upper bound data
		for _, data := range namespace.GetPredictedUpperboundData() {
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
				namespacePrediction.AddUpperBoundSample(metricType, granularity, sample)
			}
		}

		// Handle predicted lower bound data
		for _, data := range namespace.GetPredictedLowerboundData() {
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
				namespacePrediction.AddLowerBoundSample(metricType, granularity, sample)
			}
		}

		namespacePredictionMap.AddNamespacePrediction(namespacePrediction)
	}

	return namespacePredictionMap
}

type ListNamespacePredictionsRequestExtended struct {
	Request *ApiPredictions.ListNamespacePredictionsRequest
}

func (r *ListNamespacePredictionsRequestExtended) Validate() error {
	return nil
}

func (r *ListNamespacePredictionsRequestExtended) ProduceRequest() DaoPredictionTypes.ListNamespacePredictionsRequest {
	request := DaoPredictionTypes.NewListNamespacePredictionRequest()
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
