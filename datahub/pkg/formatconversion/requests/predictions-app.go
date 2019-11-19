package requests

import (
	DaoPredictionTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	Metadata "github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
	"github.com/golang/protobuf/ptypes"
)

type CreateApplicationPredictionsRequestExtended struct {
	ApiPredictions.CreateApplicationPredictionsRequest
}

func (r *CreateApplicationPredictionsRequestExtended) Validate() error {
	return nil
}

func (r *CreateApplicationPredictionsRequestExtended) ProducePredictions() DaoPredictionTypes.ApplicationPredictionMap {
	applicationPredictionMap := DaoPredictionTypes.NewApplicationPredictionMap()

	for _, application := range r.GetApplicationPredictions() {
		// Normalize request
		objectMeta := NewObjectMeta(application.GetObjectMeta())
		objectMeta.NodeName = ""

		applicationPrediction := DaoPredictionTypes.NewApplicationPrediction()
		applicationPrediction.ObjectMeta = objectMeta

		// Handle predicted raw data
		for _, data := range application.GetPredictedRawData() {
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
				applicationPrediction.AddRawSample(metricType, granularity, sample)
			}
		}

		// Handle predicted upper bound data
		for _, data := range application.GetPredictedUpperboundData() {
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
				applicationPrediction.AddUpperBoundSample(metricType, granularity, sample)
			}
		}

		// Handle predicted lower bound data
		for _, data := range application.GetPredictedLowerboundData() {
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
				applicationPrediction.AddLowerBoundSample(metricType, granularity, sample)
			}
		}

		applicationPredictionMap.AddApplicationPrediction(applicationPrediction)
	}

	return applicationPredictionMap
}

type ListApplicationPredictionsRequestExtended struct {
	Request *ApiPredictions.ListApplicationPredictionsRequest
}

func (r *ListApplicationPredictionsRequestExtended) Validate() error {
	return nil
}

func (r *ListApplicationPredictionsRequestExtended) ProduceRequest() DaoPredictionTypes.ListApplicationPredictionsRequest {
	request := DaoPredictionTypes.NewListApplicationPredictionRequest()
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
