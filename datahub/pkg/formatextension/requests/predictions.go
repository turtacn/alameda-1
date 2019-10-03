package requests

import (
	DaoPrediction "github.com/containers-ai/alameda/datahub/pkg/dao/prediction"
	Metric "github.com/containers-ai/alameda/datahub/pkg/metric"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
)

type CreatePodPredictionsRequestExtended struct {
	DatahubV1alpha1.CreatePodPredictionsRequest
}

func (r *CreatePodPredictionsRequestExtended) Validate() error {
	return nil
}

func (r *CreatePodPredictionsRequestExtended) ProducePredictions() []*DaoPrediction.ContainerPrediction {
	var (
		containerPredictions []*DaoPrediction.ContainerPrediction
	)

	for _, datahubPodPrediction := range r.PodPredictions {

		podNamespace := ""
		podName := ""
		if datahubPodPrediction.GetNamespacedName() != nil {
			podNamespace = datahubPodPrediction.GetNamespacedName().GetNamespace()
			podName = datahubPodPrediction.GetNamespacedName().GetName()
		}

		for _, datahubContainerPrediction := range datahubPodPrediction.GetContainerPredictions() {
			containerName := datahubContainerPrediction.GetName()

			containerPrediction := DaoPrediction.ContainerPrediction{
				Namespace:        podNamespace,
				PodName:          podName,
				ContainerName:    containerName,
				PredictionsRaw:   make(map[Metric.ContainerMetricType][]Metric.Sample),
				PredictionsUpper: make(map[Metric.ContainerMetricType][]Metric.Sample),
				PredictionsLower: make(map[Metric.ContainerMetricType][]Metric.Sample),
			}

			r.fillMetricData(datahubContainerPrediction.GetPredictedRawData(), &containerPrediction, Metric.ContainerMetricKindRaw)
			r.fillMetricData(datahubContainerPrediction.GetPredictedUpperboundData(), &containerPrediction, Metric.ContainerMetricKindUpperbound)
			r.fillMetricData(datahubContainerPrediction.GetPredictedLowerboundData(), &containerPrediction, Metric.ContainerMetricKindLowerbound)

			containerPredictions = append(containerPredictions, &containerPrediction)
		}
	}

	return containerPredictions
}

func (r *CreatePodPredictionsRequestExtended) fillMetricData(data []*DatahubV1alpha1.MetricData, containerPrediction *DaoPrediction.ContainerPrediction, kind Metric.ContainerMetricKind) {
	for _, rawData := range data {
		samples := []Metric.Sample{}
		for _, datahubSample := range rawData.GetData() {
			time, err := ptypes.Timestamp(datahubSample.GetTime())
			if err != nil {
				scope.Error(" failed: " + err.Error())
			}
			sample := Metric.Sample{
				Timestamp: time,
				Value:     datahubSample.GetNumValue(),
			}
			samples = append(samples, sample)
		}

		var metricType Metric.ContainerMetricType
		switch rawData.GetMetricType() {
		case DatahubV1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
			metricType = Metric.TypeContainerCPUUsageSecondsPercentage
		case DatahubV1alpha1.MetricType_MEMORY_USAGE_BYTES:
			metricType = Metric.TypeContainerMemoryUsageBytes
		}

		if kind == Metric.ContainerMetricKindRaw {
			containerPrediction.PredictionsRaw[metricType] = samples
		}
		if kind == Metric.ContainerMetricKindUpperbound {
			containerPrediction.PredictionsUpper[metricType] = samples
		}
		if kind == Metric.ContainerMetricKindLowerbound {
			containerPrediction.PredictionsLower[metricType] = samples
		}
	}
}

type CreateNodePredictionsRequestExtended struct {
	DatahubV1alpha1.CreateNodePredictionsRequest
}

func (r *CreateNodePredictionsRequestExtended) Validate() error {
	return nil
}

func (r *CreateNodePredictionsRequestExtended) ProducePredictions() []*DaoPrediction.NodePrediction {
	var (
		NodePredictions []*DaoPrediction.NodePrediction
	)

	for _, datahubNodePrediction := range r.NodePredictions {

		nodeName := datahubNodePrediction.GetName()
		isScheduled := datahubNodePrediction.GetIsScheduled()

		for _, rawData := range datahubNodePrediction.GetPredictedRawData() {

			samples := []Metric.Sample{}
			for _, datahubSample := range rawData.GetData() {
				time, err := ptypes.Timestamp(datahubSample.GetTime())
				if err != nil {
					scope.Error(" failed: " + err.Error())
				}
				sample := Metric.Sample{
					Timestamp: time,
					Value:     datahubSample.GetNumValue(),
				}
				samples = append(samples, sample)
			}

			NodePrediction := DaoPrediction.NodePrediction{
				NodeName:    nodeName,
				IsScheduled: isScheduled,
				Predictions: make(map[Metric.NodeMetricType][]Metric.Sample),
			}

			var metricType Metric.ContainerMetricType
			switch rawData.GetMetricType() {
			case DatahubV1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
				metricType = Metric.TypeNodeCPUUsageSecondsPercentage
			case DatahubV1alpha1.MetricType_MEMORY_USAGE_BYTES:
				metricType = Metric.TypeNodeMemoryUsageBytes
			}
			NodePrediction.Predictions[metricType] = samples

			NodePredictions = append(NodePredictions, &NodePrediction)
		}
	}

	return NodePredictions
}

type ListPodPredictionsRequestExtended struct {
	Request *DatahubV1alpha1.ListPodPredictionsRequest
}

func (r *ListPodPredictionsRequestExtended) Validate() error {
	return nil
}

func (r *ListPodPredictionsRequestExtended) ProduceRequest() DaoPrediction.ListPodPredictionsRequest {
	var (
		namespace      string
		podName        string
		queryCondition DBCommon.QueryCondition
		granularity    int64
	)

	if r.Request.GetNamespacedName() != nil {
		namespace = r.Request.GetNamespacedName().GetNamespace()
		podName = r.Request.GetNamespacedName().GetName()
	}

	if r.Request.GetGranularity() == 0 {
		granularity = 30
	} else {
		granularity = r.Request.GetGranularity()
	}

	queryCondition = QueryConditionExtend{r.Request.GetQueryCondition()}.QueryCondition()
	listContainerPredictionsRequest := DaoPrediction.ListPodPredictionsRequest{
		Namespace:      namespace,
		PodName:        podName,
		ModelId:        r.Request.GetModelId(),
		PredictionId:   r.Request.GetPredictionId(),
		QueryCondition: queryCondition,
		Granularity:    granularity,
	}

	return listContainerPredictionsRequest
}

type ListNodePredictionsRequestExtended struct {
	Request *DatahubV1alpha1.ListNodePredictionsRequest
}

func (r *ListNodePredictionsRequestExtended) Validate() error {
	return nil
}

func (r *ListNodePredictionsRequestExtended) ProduceRequest() DaoPrediction.ListNodePredictionsRequest {
	var (
		nodeNames      []string
		queryCondition DBCommon.QueryCondition
		granularity    int64
	)

	for _, nodeName := range r.Request.GetNodeNames() {
		nodeNames = append(nodeNames, nodeName)
	}

	if r.Request.GetGranularity() == 0 {
		granularity = 30
	} else {
		granularity = r.Request.GetGranularity()
	}

	queryCondition = QueryConditionExtend{r.Request.GetQueryCondition()}.QueryCondition()
	listNodePredictionsRequest := DaoPrediction.ListNodePredictionsRequest{
		NodeNames:      nodeNames,
		ModelId:        r.Request.GetModelId(),
		PredictionId:   r.Request.GetPredictionId(),
		Granularity:    granularity,
		QueryCondition: queryCondition,
	}

	return listNodePredictionsRequest
}
