package requests

import (
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/metric/types"
	Metric "github.com/containers-ai/alameda/datahub/pkg/metric"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
)

var ContainerMetricTypeNameMap = map[DatahubV1alpha1.MetricType]Metric.ContainerMetricType{
	DatahubV1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE: Metric.TypeContainerCPUUsageSecondsPercentage,
	DatahubV1alpha1.MetricType_MEMORY_USAGE_BYTES:           Metric.TypeContainerMemoryUsageBytes,
}

var NodeMetricTypeNameMap = map[DatahubV1alpha1.MetricType]Metric.NodeMetricType{
	DatahubV1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE: Metric.TypeNodeCPUUsageSecondsPercentage,
	DatahubV1alpha1.MetricType_MEMORY_USAGE_BYTES:           Metric.TypeNodeMemoryUsageBytes,
}

type CreateNodeMetricsRequestExtended struct {
	DatahubV1alpha1.CreateNodeMetricsRequest
}

func (r *CreateNodeMetricsRequestExtended) Validate() error {
	return nil
}

func (r *CreateNodeMetricsRequestExtended) ProduceMetrics() DaoMetricTypes.NodeMetricMap {
	nodeMetricMap := DaoMetricTypes.NewNodeMetricMap()

	for _, node := range r.GetNodeMetrics() {
		nodeMetric := DaoMetricTypes.NewNodeMetric()
		nodeMetric.NodeName = node.GetName()

		for _, data := range node.GetMetricData() {
			metricType := NodeMetricTypeNameMap[data.GetMetricType()]
			for _, sample := range data.GetData() {
				timestamp, err := ptypes.Timestamp(sample.GetTime())
				if err != nil {
					scope.Error(" failed: " + err.Error())
				}
				sample := Metric.Sample{
					Timestamp: timestamp,
					Value:     sample.GetNumValue(),
				}
				nodeMetric.AddSample(metricType, sample)
			}
		}

		nodeMetricMap.AddNodeMetric(nodeMetric)
	}

	return nodeMetricMap
}

type CreatePodMetricsRequestExtended struct {
	DatahubV1alpha1.CreatePodMetricsRequest
}

func (r *CreatePodMetricsRequestExtended) Validate() error {
	return nil
}

func (r *CreatePodMetricsRequestExtended) ProduceMetrics() DaoMetricTypes.PodMetricMap {
	podMetricMap := DaoMetricTypes.NewPodMetricMap()

	rateRange := int64(5)
	if r.GetRateRange() != 0 {
		rateRange = int64(r.GetRateRange())
	}

	for _, pod := range r.GetPodMetrics() {
		namespace := pod.GetNamespacedName().GetNamespace()
		podName := pod.GetNamespacedName().GetName()

		podMetric := DaoMetricTypes.NewPodMetric()
		podMetric.Namespace = namespace
		podMetric.PodName = podName
		podMetric.RateRange = rateRange

		for _, container := range pod.GetContainerMetrics() {
			containerName := container.GetName()

			containerMetric := DaoMetricTypes.NewContainerMetric()
			containerMetric.Namespace = namespace
			containerMetric.PodName = podName
			containerMetric.ContainerName = containerName
			containerMetric.RateRange = rateRange

			for _, data := range container.GetMetricData() {
				metricType := ContainerMetricTypeNameMap[data.GetMetricType()]
				for _, sample := range data.GetData() {
					timestamp, err := ptypes.Timestamp(sample.GetTime())
					if err != nil {
						scope.Error(" failed: " + err.Error())
					}
					sample := Metric.Sample{
						Timestamp: timestamp,
						Value:     sample.GetNumValue(),
					}
					containerMetric.AddSample(metricType, sample)
				}
			}

			podMetric.ContainerMetricMap.MetricMap[containerName] = containerMetric
		}

		podMetricMap.MetricMap[podName] = podMetric
	}

	return podMetricMap
}

type ListNodeMetricsRequestExtended struct {
	Request *DatahubV1alpha1.ListNodeMetricsRequest
}

func (r *ListNodeMetricsRequestExtended) Validate() error {
	return nil
}

func (r *ListNodeMetricsRequestExtended) ProduceRequest() DaoMetricTypes.ListNodeMetricsRequest {
	nodeNames := r.Request.GetNodeNames()

	queryCondition := QueryConditionExtend{r.Request.GetQueryCondition()}.QueryCondition()
	listNodeMetricsRequest := DaoMetricTypes.ListNodeMetricsRequest{
		QueryCondition: queryCondition,
		NodeNames:      nodeNames,
	}

	return listNodeMetricsRequest
}

type ListPodMetricsRequestExtended struct {
	Request *DatahubV1alpha1.ListPodMetricsRequest
}

func (r *ListPodMetricsRequestExtended) Validate() error {
	return nil
}

func (r *ListPodMetricsRequestExtended) ProduceRequest() DaoMetricTypes.ListPodMetricsRequest {
	namespace := ""
	podName := ""
	rateRange := int64(5)

	if r.Request.GetNamespacedName() != nil {
		namespace = r.Request.GetNamespacedName().GetNamespace()
		podName = r.Request.GetNamespacedName().GetName()
	}

	if r.Request.GetRateRange() != 0 {
		rateRange = int64(r.Request.GetRateRange())
	}

	queryCondition := QueryConditionExtend{r.Request.GetQueryCondition()}.QueryCondition()
	listPodMetricsRequest := DaoMetricTypes.ListPodMetricsRequest{
		QueryCondition: queryCondition,
		Namespace:      namespace,
		PodName:        podName,
		RateRange:      rateRange,
	}

	return listPodMetricsRequest
}
