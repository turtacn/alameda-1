package datahub

import (
	DaoPrediction "github.com/containers-ai/alameda/datahub/pkg/dao/prediction"
	DaoScore "github.com/containers-ai/alameda/datahub/pkg/dao/score"
	Metric "github.com/containers-ai/alameda/datahub/pkg/metric"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
	"time"
)

type datahubListPodMetricsRequestExtended struct {
	datahub_v1alpha1.ListPodMetricsRequest
}

func (r datahubListPodMetricsRequestExtended) validate() error {
	return nil
}

type datahubListNodeMetricsRequestExtended struct {
	datahub_v1alpha1.ListNodeMetricsRequest
}

func (r datahubListNodeMetricsRequestExtended) validate() error {
	return nil
}

type datahubCreatePodPredictionsRequestExtended struct {
	datahub_v1alpha1.CreatePodPredictionsRequest
}

func (r datahubCreatePodPredictionsRequestExtended) validate() error {
	return nil
}

func (r datahubCreatePodPredictionsRequestExtended) daoContainerPredictions() []*DaoPrediction.ContainerPrediction {

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

func (r datahubCreatePodPredictionsRequestExtended) fillMetricData(data []*datahub_v1alpha1.MetricData, containerPrediction *DaoPrediction.ContainerPrediction, kind Metric.ContainerMetricKind) {
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
		case datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
			metricType = Metric.TypeContainerCPUUsageSecondsPercentage
		case datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES:
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

type datahubCreateNodePredictionsRequestExtended struct {
	datahub_v1alpha1.CreateNodePredictionsRequest
}

func (r datahubCreateNodePredictionsRequestExtended) validate() error {
	return nil
}

func (r datahubCreateNodePredictionsRequestExtended) daoNodePredictions() []*DaoPrediction.NodePrediction {

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
			case datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE:
				metricType = Metric.TypeNodeCPUUsageSecondsPercentage
			case datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES:
				metricType = Metric.TypeNodeMemoryUsageBytes
			}
			NodePrediction.Predictions[metricType] = samples

			NodePredictions = append(NodePredictions, &NodePrediction)
		}
	}

	return NodePredictions
}

type datahubListPodPredictionsRequestExtended struct {
	request *datahub_v1alpha1.ListPodPredictionsRequest
}

func (r datahubListPodPredictionsRequestExtended) daoListPodPredictionsRequest() DaoPrediction.ListPodPredictionsRequest {

	var (
		namespace      string
		podName        string
		queryCondition DBCommon.QueryCondition
		granularity    int64
	)

	if r.request.GetNamespacedName() != nil {
		namespace = r.request.GetNamespacedName().GetNamespace()
		podName = r.request.GetNamespacedName().GetName()
	}

	if r.request.GetGranularity() == 0 {
		granularity = 30
	} else {
		granularity = r.request.GetGranularity()
	}

	queryCondition = datahubQueryConditionExtend{r.request.GetQueryCondition()}.daoQueryCondition()
	listContainerPredictionsRequest := DaoPrediction.ListPodPredictionsRequest{
		Namespace:      namespace,
		PodName:        podName,
		QueryCondition: queryCondition,
		Granularity:    granularity,
	}

	return listContainerPredictionsRequest
}

type datahubListNodePredictionsRequestExtended struct {
	request *datahub_v1alpha1.ListNodePredictionsRequest
}

func (r datahubListNodePredictionsRequestExtended) daoListNodePredictionsRequest() DaoPrediction.ListNodePredictionsRequest {

	var (
		nodeNames      []string
		queryCondition DBCommon.QueryCondition
		granularity    int64
	)

	for _, nodeName := range r.request.GetNodeNames() {
		nodeNames = append(nodeNames, nodeName)
	}

	if r.request.GetGranularity() == 0 {
		granularity = 30
	} else {
		granularity = r.request.GetGranularity()
	}

	queryCondition = datahubQueryConditionExtend{r.request.GetQueryCondition()}.daoQueryCondition()
	listNodePredictionsRequest := DaoPrediction.ListNodePredictionsRequest{
		NodeNames:      nodeNames,
		QueryCondition: queryCondition,
		Granularity:    granularity,
	}

	return listNodePredictionsRequest
}

type datahubListSimulatedSchedulingScoresRequestExtended struct {
	request *datahub_v1alpha1.ListSimulatedSchedulingScoresRequest
}

func (r datahubListSimulatedSchedulingScoresRequestExtended) daoLisRequest() DaoScore.ListRequest {

	var (
		queryCondition DBCommon.QueryCondition
	)

	queryCondition = datahubQueryConditionExtend{r.request.GetQueryCondition()}.daoQueryCondition()
	listRequest := DaoScore.ListRequest{
		QueryCondition: queryCondition,
	}

	return listRequest
}

var (
	datahubAggregateFunction_DAOAggregateFunction = map[datahub_v1alpha1.TimeRange_AggregateFunction]DBCommon.AggregateFunction{
		datahub_v1alpha1.TimeRange_NONE: DBCommon.None,
		datahub_v1alpha1.TimeRange_MAX:  DBCommon.MaxOverTime,
	}
)

type datahubQueryConditionExtend struct {
	queryCondition *datahub_v1alpha1.QueryCondition
}

func (d datahubQueryConditionExtend) daoQueryCondition() DBCommon.QueryCondition {

	var (
		queryStartTime      *time.Time
		queryEndTime        *time.Time
		queryStepTime       *time.Duration
		queryTimestampOrder int
		queryLimit          int
		queryCondition      = DBCommon.QueryCondition{}
		aggregateFunc       = DBCommon.None
	)

	if d.queryCondition == nil {
		return queryCondition
	}

	if d.queryCondition.GetTimeRange() != nil {
		timeRange := d.queryCondition.GetTimeRange()
		if timeRange.GetStartTime() != nil {
			tmpTime, _ := ptypes.Timestamp(timeRange.GetStartTime())
			queryStartTime = &tmpTime
		}
		if timeRange.GetEndTime() != nil {
			tmpTime, _ := ptypes.Timestamp(timeRange.GetEndTime())
			queryEndTime = &tmpTime
		}
		if timeRange.GetStep() != nil {
			tmpTime, _ := ptypes.Duration(timeRange.GetStep())
			queryStepTime = &tmpTime
		}

		switch d.queryCondition.GetOrder() {
		case datahub_v1alpha1.QueryCondition_ASC:
			queryTimestampOrder = DBCommon.Asc
		case datahub_v1alpha1.QueryCondition_DESC:
			queryTimestampOrder = DBCommon.Desc
		default:
			queryTimestampOrder = DBCommon.Asc
		}

		queryLimit = int(d.queryCondition.GetLimit())
	}
	queryTimestampOrder = int(d.queryCondition.GetOrder())
	queryLimit = int(d.queryCondition.GetLimit())

	if aggFunc, exist := datahubAggregateFunction_DAOAggregateFunction[datahub_v1alpha1.TimeRange_AggregateFunction(d.queryCondition.TimeRange.AggregateFunction)]; exist {
		aggregateFunc = aggFunc
	}

	queryCondition = DBCommon.QueryCondition{
		StartTime:                 queryStartTime,
		EndTime:                   queryEndTime,
		StepTime:                  queryStepTime,
		TimestampOrder:            queryTimestampOrder,
		Limit:                     queryLimit,
		AggregateOverTimeFunction: aggregateFunc,
	}
	return queryCondition
}
