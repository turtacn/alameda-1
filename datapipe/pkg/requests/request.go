package requests

import (
	"github.com/golang/protobuf/ptypes"
	"time"

	"github.com/containers-ai/alameda/datapipe/pkg/dao"
	CommonAPI "github.com/containers-ai/api/common"
	dataPipeMetricsAPI "github.com/containers-ai/api/datapipe/metrics"
	dataPipePredictionsAPI "github.com/containers-ai/api/datapipe/predictions"
)

type DatahubListPodMetricsRequestExtended struct {
	dataPipeMetricsAPI.ListPodMetricsRequest
}

func (r DatahubListPodMetricsRequestExtended) Validate() error {
	return nil
}

type DatahubListNodeMetricsRequestExtended struct {
	dataPipeMetricsAPI.ListNodeMetricsRequest
}

func (r DatahubListNodeMetricsRequestExtended) Validate() error {
	return nil
}

type DatahubListPodPredictionsRequestExtended struct {
	request *dataPipePredictionsAPI.ListPodPredictionsRequest
}

var (
	DatahubAggregateFunction_DAOAggregateFunction = map[CommonAPI.TimeRange_AggregateFunction]dao.AggregateFunction{
		CommonAPI.TimeRange_NONE: dao.None,
		CommonAPI.TimeRange_MAX:  dao.Max,
	}
)

type DatahubQueryConditionExtend struct {
	QueryCondition *CommonAPI.QueryCondition
}

func (d DatahubQueryConditionExtend) DaoQueryCondition() dao.QueryCondition {

	var (
		queryStartTime      *time.Time
		queryEndTime        *time.Time
		queryStepTime       *time.Duration
		queryTimestampOrder int
		queryLimit          int
		queryCondition      = dao.QueryCondition{}
		aggregateFunc       = dao.None
	)

	if d.QueryCondition == nil {
		return queryCondition
	}

	if d.QueryCondition.GetTimeRange() != nil {
		timeRange := d.QueryCondition.GetTimeRange()
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

		switch d.QueryCondition.GetOrder() {
		case CommonAPI.QueryCondition_ASC:
			queryTimestampOrder = dao.Asc
		case CommonAPI.QueryCondition_DESC:
			queryTimestampOrder = dao.Desc
		default:
			queryTimestampOrder = dao.Asc
		}

		queryLimit = int(d.QueryCondition.GetLimit())
	}
	queryTimestampOrder = int(d.QueryCondition.GetOrder())
	queryLimit = int(d.QueryCondition.GetLimit())

	if aggFunc, exist := DatahubAggregateFunction_DAOAggregateFunction[CommonAPI.TimeRange_AggregateFunction(d.QueryCondition.TimeRange.AggregateFunction)]; exist {
		aggregateFunc = aggFunc
	}

	queryCondition = dao.QueryCondition{
		StartTime:      queryStartTime,
		EndTime:        queryEndTime,
		StepTime:       queryStepTime,
		TimestampOrder: queryTimestampOrder,
		Limit:          queryLimit,
		AggregateOverTimeFunction: aggregateFunc,
	}
	return queryCondition
}
