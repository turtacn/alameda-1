package requests

import (
	DaoPredictionTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
	"github.com/golang/protobuf/ptypes"
)

type CreateClusterPredictionsRequestExtended struct {
	ApiPredictions.CreateClusterPredictionsRequest
}

func (r *CreateClusterPredictionsRequestExtended) Validate() error {
	return nil
}

func (r *CreateClusterPredictionsRequestExtended) ProducePredictions() DaoPredictionTypes.ClusterPredictionMap {
	clusterPredictionMap := DaoPredictionTypes.NewClusterPredictionMap()

	for _, cluster := range r.GetClusterPredictions() {
		clusterPrediction := DaoPredictionTypes.NewClusterPrediction()
		clusterPrediction.ObjectMeta.Name = cluster.GetObjectMeta().GetName()

		// Handle predicted raw data
		for _, data := range cluster.GetPredictedRawData() {
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
				clusterPrediction.AddRawSample(metricType, granularity, sample)
			}
		}

		// Handle predicted upper bound data
		for _, data := range cluster.GetPredictedUpperboundData() {
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
				clusterPrediction.AddUpperBoundSample(metricType, granularity, sample)
			}
		}

		// Handle predicted lower bound data
		for _, data := range cluster.GetPredictedLowerboundData() {
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
				clusterPrediction.AddLowerBoundSample(metricType, granularity, sample)
			}
		}

		clusterPredictionMap.AddClusterPrediction(clusterPrediction)
	}

	return clusterPredictionMap
}

type ListClusterPredictionsRequestExtended struct {
	Request *ApiPredictions.ListClusterPredictionsRequest
}

func (r *ListClusterPredictionsRequestExtended) Validate() error {
	return nil
}

func (r *ListClusterPredictionsRequestExtended) ProduceRequest() DaoPredictionTypes.ListClusterPredictionsRequest {
	request := DaoPredictionTypes.NewListClusterPredictionRequest()
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
