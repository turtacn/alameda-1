package requests

import (
	FormatConvert "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	"github.com/golang/protobuf/ptypes"
	"time"
)

type QueryConditionExtend struct {
	Condition *ApiCommon.QueryCondition
}

func (d QueryConditionExtend) QueryCondition() DBCommon.QueryCondition {
	var (
		queryStartTime      *time.Time
		queryEndTime        *time.Time
		queryStepTime       *time.Duration
		queryTimestampOrder int
		queryLimit          int
		queryCondition      = DBCommon.QueryCondition{}
		aggregateFunc       = DBCommon.None
	)

	if d.Condition == nil {
		return queryCondition
	}

	if d.Condition.GetTimeRange() != nil {
		timeRange := d.Condition.GetTimeRange()
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

		switch d.Condition.GetOrder() {
		case ApiCommon.QueryCondition_ASC:
			queryTimestampOrder = DBCommon.Asc
		case ApiCommon.QueryCondition_DESC:
			queryTimestampOrder = DBCommon.Desc
		default:
			queryTimestampOrder = DBCommon.Asc
		}

		queryLimit = int(d.Condition.GetLimit())
	}
	queryTimestampOrder = int(d.Condition.GetOrder())
	queryLimit = int(d.Condition.GetLimit())

	if aggFunc, exist := FormatConvert.AggregateFunctionNameMap[ApiCommon.TimeRange_AggregateFunction(d.Condition.TimeRange.AggregateFunction)]; exist {
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
