package requests

import (
	DaoPredictionTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
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
		nodePrediction := DaoPredictionTypes.NewNodePrediction()
		nodePrediction.IsScheduled = node.GetIsScheduled()
		nodePrediction.ObjectMeta.Name = node.GetObjectMeta().GetName()

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

type CreatePodPredictionsRequestExtended struct {
	ApiPredictions.CreatePodPredictionsRequest
}

func (r *CreatePodPredictionsRequestExtended) Validate() error {
	return nil
}

func (r *CreatePodPredictionsRequestExtended) ProducePredictions() DaoPredictionTypes.PodPredictionMap {
	podPredictionMap := DaoPredictionTypes.NewPodPredictionMap()

	for _, pod := range r.GetPodPredictions() {
		namespace := pod.GetObjectMeta().GetNamespace()
		podName := pod.GetObjectMeta().GetName()

		podPrediction := DaoPredictionTypes.NewPodPrediction()
		podPrediction.ObjectMeta.Namespace = namespace
		podPrediction.ObjectMeta.Name = podName

		for _, container := range pod.GetContainerPredictions() {
			containerName := container.GetName()

			containerPrediction := DaoPredictionTypes.NewContainerPrediction()
			containerPrediction.Namespace = namespace
			containerPrediction.PodName = podName
			containerPrediction.ContainerName = containerName

			// Handle predicted raw data
			for _, data := range container.GetPredictedRawData() {
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
					containerPrediction.AddRawSample(metricType, granularity, sample)
				}
			}

			// Handle predicted upper bound data
			for _, data := range container.GetPredictedUpperboundData() {
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
					containerPrediction.AddUpperBoundSample(metricType, granularity, sample)
				}
			}

			// Handle predicted lower bound data
			for _, data := range container.GetPredictedLowerboundData() {
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
					containerPrediction.AddLowerBoundSample(metricType, granularity, sample)
				}
			}

			podPrediction.ContainerPredictionMap.AddContainerPrediction(containerPrediction)
		}

		podPredictionMap.AddPodPrediction(podPrediction)
	}

	return podPredictionMap
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

type ListPodPredictionsRequestExtended struct {
	Request *ApiPredictions.ListPodPredictionsRequest
}

func (r *ListPodPredictionsRequestExtended) Validate() error {
	return nil
}

func (r *ListPodPredictionsRequestExtended) ProduceRequest() DaoPredictionTypes.ListPodPredictionsRequest {
	request := DaoPredictionTypes.NewListPodPredictionsRequest()
	request.QueryCondition = QueryConditionExtend{r.Request.GetQueryCondition()}.QueryCondition()
	request.ModelId = r.Request.GetModelId()
	request.PredictionId = r.Request.GetPredictionId()
	request.FillDays = r.Request.GetFillDays()
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
