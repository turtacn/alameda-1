package datahub

import (
	"errors"
	"time"

	prediction_dao "github.com/containers-ai/alameda/datahub/pkg/dao/prediction"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
)

type datahubListPodMetricsRequestExtended struct {
	datahub_v1alpha1.ListPodMetricsRequest
}

func (r datahubListPodMetricsRequestExtended) validate() error {

	var (
		startTime *timestamp.Timestamp
		endTime   *timestamp.Timestamp
	)

	if r.TimeRange == nil {
		return errors.New("field \"time_range\" cannot be empty")
	}

	startTime = r.TimeRange.StartTime
	endTime = r.TimeRange.EndTime
	if startTime == nil || endTime == nil {
		return errors.New("field \"start_time\" and \"end_time\"  cannot be empty")
	}

	if startTime.Seconds+int64(startTime.Nanos) >= endTime.Seconds+int64(endTime.Nanos) {
		return errors.New("\"end_time\" must not be before \"start_time\"")
	}

	return nil
}

type datahubListNodeMetricsRequestExtended struct {
	datahub_v1alpha1.ListNodeMetricsRequest
}

func (r datahubListNodeMetricsRequestExtended) validate() error {

	var (
		startTime *timestamp.Timestamp
		endTime   *timestamp.Timestamp
	)

	if r.TimeRange == nil {
		return errors.New("field \"time_range\" cannot be empty")
	}

	startTime = r.TimeRange.StartTime
	endTime = r.TimeRange.EndTime
	if startTime == nil || endTime == nil {
		return errors.New("field \"start_time\" and \"end_time\"  cannot be empty")
	}

	if startTime.Seconds+int64(startTime.Nanos) >= endTime.Seconds+int64(endTime.Nanos) {
		return errors.New("\"end_time\" must not be before \"start_time\"")
	}

	return nil
}

type datahubCreatePodPredictionsRequestExtended struct {
	datahub_v1alpha1.CreatePodPredictionsRequest
}

func (r datahubCreatePodPredictionsRequestExtended) validate() error {
	return nil
}

func (r datahubCreatePodPredictionsRequestExtended) daoContainerPredictions() []*prediction_dao.ContainerPrediction {

	var (
		containerPredictions []*prediction_dao.ContainerPrediction
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

			for _, rawData := range datahubContainerPrediction.GetPredictedRawData() {

				containerPrediction := prediction_dao.ContainerPrediction{
					Namespace:     podNamespace,
					PodName:       podName,
					ContainerName: containerName,
				}

				samples := []prediction_dao.Sample{}
				for _, datahubSample := range rawData.GetData() {
					time, err := ptypes.Timestamp(datahubSample.GetTime())
					if err != nil {
						scope.Error(" failed: " + err.Error())
					}
					sample := prediction_dao.Sample{
						Timestamp: time,
						Value:     datahubSample.GetNumValue(),
					}
					samples = append(samples, sample)
				}

				metricType := rawData.GetMetricType()
				switch metricType {
				case datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
					containerPrediction.CPUPredictions = samples
				case datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES:
					containerPrediction.MemoryPredictions = samples
				}

				containerPredictions = append(containerPredictions, &containerPrediction)
			}
		}
	}

	return containerPredictions
}

type datahubCreateNodePredictionsRequestExtended struct {
	datahub_v1alpha1.CreateNodePredictionsRequest
}

func (r datahubCreateNodePredictionsRequestExtended) validate() error {
	return nil
}

func (r datahubCreateNodePredictionsRequestExtended) daoNodePredictions() []*prediction_dao.NodePrediction {

	var (
		NodePredictions []*prediction_dao.NodePrediction
	)

	for _, datahubNodePrediction := range r.NodePredictions {

		nodeName := datahubNodePrediction.GetName()
		isScheduled := datahubNodePrediction.GetIsScheduled()

		for _, rawData := range datahubNodePrediction.GetPredictedRawData() {

			samples := []prediction_dao.Sample{}
			for _, datahubSample := range rawData.GetData() {
				time, err := ptypes.Timestamp(datahubSample.GetTime())
				if err != nil {
					scope.Error(" failed: " + err.Error())
				}
				sample := prediction_dao.Sample{
					Timestamp: time,
					Value:     datahubSample.GetNumValue(),
				}
				samples = append(samples, sample)
			}

			NodePrediction := prediction_dao.NodePrediction{
				NodeName:    nodeName,
				IsScheduled: isScheduled,
			}

			metricType := rawData.GetMetricType()
			switch metricType {
			case datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
				NodePrediction.CPUUsagePredictions = samples
			case datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES:
				NodePrediction.MemoryUsagePredictions = samples
			}

			NodePredictions = append(NodePredictions, &NodePrediction)
		}
	}

	return NodePredictions
}

type datahubListPodPredictionsRequestExtended struct {
	datahub_v1alpha1.ListPodPredictionsRequest
}

func (r datahubListPodPredictionsRequestExtended) daoListPodPredictionsRequest() prediction_dao.ListPodPredictionsRequest {

	var (
		namespace string
		podName   string
		startTime *time.Time
		endTime   *time.Time
	)

	if r.GetNamespacedName() != nil {
		namespace = r.GetNamespacedName().GetNamespace()
		podName = r.GetNamespacedName().GetName()
	}

	if r.GetTimeRange() != nil {

		if r.GetTimeRange().GetStartTime() != nil {
			tmpStartTime, _ := ptypes.Timestamp(r.GetTimeRange().GetStartTime())
			startTime = &tmpStartTime
		}

		if r.GetTimeRange().GetEndTime() != nil {
			tmpEndTime, _ := ptypes.Timestamp(r.GetTimeRange().GetEndTime())
			endTime = &tmpEndTime
		}
	}

	listContainerPredictionsRequest := prediction_dao.ListPodPredictionsRequest{
		Namespace: namespace,
		PodName:   podName,
		StartTime: startTime,
		EndTime:   endTime,
	}

	return listContainerPredictionsRequest
}

type datahubListNodePredictionsRequestExtended struct {
	datahub_v1alpha1.ListNodePredictionsRequest
}

func (r datahubListNodePredictionsRequestExtended) daoListNodePredictionsRequest() prediction_dao.ListNodePredictionsRequest {

	var (
		nodeNames []string
		startTime *time.Time
		endTime   *time.Time
	)

	if r.GetTimeRange() != nil {

		if r.GetTimeRange().GetStartTime() != nil {
			tmpStartTime, _ := ptypes.Timestamp(r.GetTimeRange().GetStartTime())
			startTime = &tmpStartTime
		}

		if r.GetTimeRange().GetEndTime() != nil {
			tmpEndTime, _ := ptypes.Timestamp(r.GetTimeRange().GetEndTime())
			endTime = &tmpEndTime
		}
	}

	for _, nodeName := range r.GetNodeName() {
		nodeNames = append(nodeNames, nodeName)
	}

	listNodePredictionsRequest := prediction_dao.ListNodePredictionsRequest{
		NodeNames: nodeNames,
		StartTime: startTime,
		EndTime:   endTime,
	}

	return listNodePredictionsRequest
}
