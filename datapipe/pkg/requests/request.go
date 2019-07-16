package requests

import (
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	CommonAPI "github.com/containers-ai/api/common"
	dataPipeMetricsAPI "github.com/containers-ai/api/datapipe/metrics"
	dataPipePredictionsAPI "github.com/containers-ai/api/datapipe/predictions"
	"github.com/golang/protobuf/ptypes"
	"time"
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

type DatahubQueryConditionExtend struct {
	QueryCondition *CommonAPI.QueryCondition
}

func (d DatahubQueryConditionExtend) DaoQueryCondition() DBCommon.QueryCondition {

	var (
		queryStartTime      *time.Time
		queryEndTime        *time.Time
		queryStepTime       *time.Duration
		queryTimestampOrder int
		queryLimit          int
		queryCondition      = DBCommon.QueryCondition{}
		aggregateFunc       = DBCommon.None
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
			queryTimestampOrder = DBCommon.Asc
		case CommonAPI.QueryCondition_DESC:
			queryTimestampOrder = DBCommon.Desc
		default:
			queryTimestampOrder = DBCommon.Asc
		}

		queryLimit = int(d.QueryCondition.GetLimit())
	}
	queryTimestampOrder = int(d.QueryCondition.GetOrder())
	queryLimit = int(d.QueryCondition.GetLimit())

	if aggFunc, exist := DBCommon.TimeRange2AggregationOverTime[CommonAPI.TimeRange_AggregateFunction(d.QueryCondition.TimeRange.AggregateFunction)]; exist {
		aggregateFunc = aggFunc
	}

	queryCondition = DBCommon.QueryCondition{
		StartTime:      queryStartTime,
		EndTime:        queryEndTime,
		StepTime:       queryStepTime,
		TimestampOrder: queryTimestampOrder,
		Limit:          queryLimit,
		AggregateOverTimeFunction: aggregateFunc,
	}
	return queryCondition
}
