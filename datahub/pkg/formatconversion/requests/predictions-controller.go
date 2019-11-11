package requests

import (
	DaoPredictionTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
	"github.com/golang/protobuf/ptypes"
)

type CreateControllerPredictionsRequestExtended struct {
	ApiPredictions.CreateControllerPredictionsRequest
}

func (r *CreateControllerPredictionsRequestExtended) Validate() error {
	return nil
}

func (r *CreateControllerPredictionsRequestExtended) ProducePredictions() DaoPredictionTypes.ControllerPredictionMap {
	controllerPredictionMap := DaoPredictionTypes.NewControllerPredictionMap()

	for _, controller := range r.GetControllerPredictions() {
		controllerPrediction := DaoPredictionTypes.NewControllerPrediction()
		controllerPrediction.ObjectMeta.Name = controller.GetObjectMeta().GetName()
		controllerPrediction.CtlKind = controller.GetKind().String()

		// Handle predicted raw data
		for _, data := range controller.GetPredictedRawData() {
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
				controllerPrediction.AddRawSample(metricType, granularity, sample)
			}
		}

		// Handle predicted upper bound data
		for _, data := range controller.GetPredictedUpperboundData() {
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
				controllerPrediction.AddUpperBoundSample(metricType, granularity, sample)
			}
		}

		// Handle predicted lower bound data
		for _, data := range controller.GetPredictedLowerboundData() {
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
				controllerPrediction.AddLowerBoundSample(metricType, granularity, sample)
			}
		}

		controllerPredictionMap.AddControllerPrediction(controllerPrediction)
	}

	return controllerPredictionMap
}

type ListControllerPredictionsRequestExtended struct {
	Request *ApiPredictions.ListControllerPredictionsRequest
}

func (r *ListControllerPredictionsRequestExtended) Validate() error {
	return nil
}

func (r *ListControllerPredictionsRequestExtended) ProduceRequest() DaoPredictionTypes.ListControllerPredictionsRequest {
	request := DaoPredictionTypes.NewListControllerPredictionRequest()
	request.QueryCondition = QueryConditionExtend{r.Request.GetQueryCondition()}.QueryCondition()
	request.ModelId = r.Request.GetModelId()
	request.PredictionId = r.Request.GetPredictionId()
	request.Granularity = 30
	if r.Request.GetGranularity() != 0 {
		request.Granularity = r.Request.GetGranularity()
	}
	if r.Request.GetObjectMeta() != nil {
		for _, objectMeta := range r.Request.GetObjectMeta() {
			request.ObjectMeta = append(request.ObjectMeta, NewObjectMeta(objectMeta))
		}
	}
	return request
}
