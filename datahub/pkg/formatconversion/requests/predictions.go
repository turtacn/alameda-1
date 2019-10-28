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
		nodePrediction.NodeName = node.GetName()

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
		namespace := pod.GetNamespacedName().GetNamespace()
		podName := pod.GetNamespacedName().GetName()

		podPrediction := DaoPredictionTypes.NewPodPrediction()
		podPrediction.Namespace = namespace
		podPrediction.PodName = podName

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
	nodeNames := make([]string, 0)
	granularity := int64(30)

	if r.Request.GetNodeNames() != nil {
		for _, nodeName := range r.Request.GetNodeNames() {
			nodeNames = append(nodeNames, nodeName)
		}
	}

	if r.Request.GetGranularity() != 0 {
		granularity = r.Request.GetGranularity()
	}

	queryCondition := QueryConditionExtend{r.Request.GetQueryCondition()}.QueryCondition()
	listNodePredictionsRequest := DaoPredictionTypes.ListNodePredictionsRequest{
		QueryCondition: queryCondition,
		NodeNames:      nodeNames,
		ModelId:        r.Request.GetModelId(),
		PredictionId:   r.Request.GetPredictionId(),
		Granularity:    granularity,
	}

	return listNodePredictionsRequest
}

type ListPodPredictionsRequestExtended struct {
	Request *ApiPredictions.ListPodPredictionsRequest
}

func (r *ListPodPredictionsRequestExtended) Validate() error {
	return nil
}

func (r *ListPodPredictionsRequestExtended) ProduceRequest() DaoPredictionTypes.ListPodPredictionsRequest {
	namespace := ""
	podName := ""
	granularity := int64(30)

	if r.Request.GetNamespacedName() != nil {
		namespace = r.Request.GetNamespacedName().GetNamespace()
		podName = r.Request.GetNamespacedName().GetName()
	}

	if r.Request.GetGranularity() != 0 {
		granularity = r.Request.GetGranularity()
	}

	queryCondition := QueryConditionExtend{r.Request.GetQueryCondition()}.QueryCondition()
	listContainerPredictionsRequest := DaoPredictionTypes.ListPodPredictionsRequest{
		QueryCondition: queryCondition,
		Namespace:      namespace,
		PodName:        podName,
		ModelId:        r.Request.GetModelId(),
		PredictionId:   r.Request.GetPredictionId(),
		Granularity:    granularity,
		FillDays:       r.Request.GetFillDays(),
	}

	return listContainerPredictionsRequest
}
