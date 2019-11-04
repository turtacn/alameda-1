package requests

import (
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiMetrics "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/metrics"
	"github.com/golang/protobuf/ptypes"
)

var MetricTypeNameMap = map[ApiCommon.MetricType]FormatEnum.MetricType{
	ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE: FormatEnum.MetricTypeCPUUsageSecondsPercentage,
	ApiCommon.MetricType_MEMORY_USAGE_BYTES:           FormatEnum.MetricTypeMemoryUsageBytes,
	ApiCommon.MetricType_POWER_USAGE_WATTS:            FormatEnum.MetricTypePowerUsageWatts,
	ApiCommon.MetricType_TEMPERATURE_CELSIUS:          FormatEnum.MetricTypeTemperatureCelsius,
	ApiCommon.MetricType_DUTY_CYCLE:                   FormatEnum.MetricTypeDutyCycle,
}

type CreateNodeMetricsRequestExtended struct {
	ApiMetrics.CreateNodeMetricsRequest
}

func (r *CreateNodeMetricsRequestExtended) Validate() error {
	return nil
}

func (r *CreateNodeMetricsRequestExtended) ProduceMetrics() DaoMetricTypes.NodeMetricMap {
	nodeMetricMap := DaoMetricTypes.NewNodeMetricMap()

	for _, node := range r.GetNodeMetrics() {
		nodeMetric := DaoMetricTypes.NewNodeMetric()
		nodeMetric.ObjectMeta = NewObjectMeta(node.GetObjectMeta())

		for _, data := range node.GetMetricData() {
			metricType := MetricTypeNameMap[data.GetMetricType()]
			for _, sample := range data.GetData() {
				timestamp, err := ptypes.Timestamp(sample.GetTime())
				if err != nil {
					scope.Error(" failed: " + err.Error())
				}
				sample := FormatTypes.Sample{
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
	ApiMetrics.CreatePodMetricsRequest
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
		podMetric := DaoMetricTypes.NewPodMetric()
		podMetric.ObjectMeta = NewObjectMeta(pod.GetObjectMeta())
		podMetric.RateRange = rateRange

		for _, container := range pod.GetContainerMetrics() {
			containerMetric := DaoMetricTypes.NewContainerMetric()
			containerMetric.Namespace = podMetric.ObjectMeta.Namespace
			containerMetric.PodName = podMetric.ObjectMeta.Name
			containerMetric.ContainerName = container.GetName()
			containerMetric.RateRange = rateRange

			for _, data := range container.GetMetricData() {
				metricType := MetricTypeNameMap[data.GetMetricType()]
				for _, sample := range data.GetData() {
					timestamp, err := ptypes.Timestamp(sample.GetTime())
					if err != nil {
						scope.Error(" failed: " + err.Error())
					}
					sample := FormatTypes.Sample{
						Timestamp: timestamp,
						Value:     sample.GetNumValue(),
					}
					containerMetric.AddSample(metricType, sample)
				}
			}

			podMetric.ContainerMetricMap.AddContainerMetric(containerMetric)
		}

		podMetricMap.AddPodMetric(podMetric)
	}

	return podMetricMap
}

type ListNodeMetricsRequestExtended struct {
	Request *ApiMetrics.ListNodeMetricsRequest
}

func (r *ListNodeMetricsRequestExtended) Validate() error {
	return nil
}

func (r *ListNodeMetricsRequestExtended) ProduceRequest() DaoMetricTypes.ListNodeMetricsRequest {
	request := DaoMetricTypes.NewListNodeMetricsRequest()
	request.QueryCondition = QueryConditionExtend{r.Request.GetQueryCondition()}.QueryCondition()
	if r.Request.GetObjectMeta() != nil {
		for _, objectMeta := range r.Request.GetObjectMeta() {
			request.ObjectMeta = append(request.ObjectMeta, NewObjectMeta(objectMeta))
		}
	}
	return request
}

type ListPodMetricsRequestExtended struct {
	Request *ApiMetrics.ListPodMetricsRequest
}

func (r *ListPodMetricsRequestExtended) Validate() error {
	return nil
}

func (r *ListPodMetricsRequestExtended) ProduceRequest() DaoMetricTypes.ListPodMetricsRequest {
	request := DaoMetricTypes.NewListPodMetricsRequest()
	request.QueryCondition = QueryConditionExtend{r.Request.GetQueryCondition()}.QueryCondition()
	request.RateRange = 5
	if r.Request.GetRateRange() != 0 {
		request.RateRange = int64(r.Request.GetRateRange())
	}
	if r.Request.GetObjectMeta() != nil {
		for _, objectMeta := range r.Request.GetObjectMeta() {
			request.ObjectMeta = append(request.ObjectMeta, NewObjectMeta(objectMeta))
		}
	}
	return request
}
