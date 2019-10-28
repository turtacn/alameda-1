package common

import (
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	Common "github.com/containers-ai/api/common"
	"github.com/golang/protobuf/ptypes"
	"time"
)

// QueryCondition Others query condition
type QueryCondition struct {
	StartTime                 *time.Time
	EndTime                   *time.Time
	Timeout                   *time.Time
	StepTime                  *time.Duration
	TimestampOrder            Order
	Limit                     int
	AggregateOverTimeFunction AggregateFunction
}

func NewQueryCondition(days, hours, seconds, steps int) *QueryCondition {
	totalSeconds := (days * 24 * 60 * 60) + (hours * 60) + seconds
	endTime      := time.Now()
	startTime    := endTime.Add(time.Duration(-totalSeconds) * time.Second)
	duration     := time.Duration(steps) * time.Second

	queryCondition := &QueryCondition{
		StartTime:                 &startTime,
		EndTime:                   &endTime,
		StepTime:                  &duration,
		TimestampOrder:            Desc,
		Limit:                     1,
		AggregateOverTimeFunction: None,
	}

	return queryCondition
}

func BuildQueryCondition(condition *Common.QueryCondition) QueryCondition {
	var (
		queryStartTime      *time.Time
		queryEndTime        *time.Time
		queryTimeout        *time.Time
		queryStepTime       *time.Duration
		queryTimestampOrder int
		queryLimit          int
		queryCondition      = QueryCondition{}
		aggregateFunc       = None
	)

	if condition == nil {
		return queryCondition
	}

	if condition.GetTimeRange() != nil {
		timeRange := condition.GetTimeRange()

		if timeRange.GetStartTime() != nil {
			tmpTime, _ := ptypes.Timestamp(timeRange.GetStartTime())
			queryStartTime = &tmpTime
		}

		if timeRange.GetEndTime() != nil {
			tmpTime, _ := ptypes.Timestamp(timeRange.GetEndTime())
			queryEndTime = &tmpTime
		}

		if timeRange.GetTimeout() != nil {
			tmpTime, _ := ptypes.Timestamp(timeRange.GetTimeout())
			queryTimeout = &tmpTime
		}

		if timeRange.GetStep() != nil {
			tmpTime, _ := ptypes.Duration(timeRange.GetStep())
			queryStepTime = &tmpTime
		}

		switch condition.GetOrder() {
		case Common.QueryCondition_ASC:
			queryTimestampOrder = Asc
		case Common.QueryCondition_DESC:
			queryTimestampOrder = Desc
		default:
			queryTimestampOrder = Asc
		}

		queryLimit = int(condition.GetLimit())
	}

	queryTimestampOrder = int(condition.GetOrder())
	queryLimit = int(condition.GetLimit())

	if aggFunc, exist := TimeRange2AggregationOverTime[Common.TimeRange_AggregateFunction(condition.GetTimeRange().GetAggregateFunction())]; exist {
		aggregateFunc = aggFunc
	}

	queryCondition = QueryCondition{
		StartTime:      queryStartTime,
		EndTime:        queryEndTime,
		Timeout:        queryTimeout,
		StepTime:       queryStepTime,
		TimestampOrder: queryTimestampOrder,
		Limit:          queryLimit,
		AggregateOverTimeFunction: aggregateFunc,
	}

	return queryCondition
}

func BuildQueryConditionV1(condition *ApiCommon.QueryCondition) *QueryCondition {
	startTime, _ := ptypes.Timestamp(condition.GetTimeRange().GetStartTime())
	endTime, _ := ptypes.Timestamp(condition.GetTimeRange().GetEndTime())
	stepTime, _ := ptypes.Duration(condition.GetTimeRange().GetStep())

	queryCondition := QueryCondition{
		StartTime: &startTime,
		EndTime: &endTime,
		StepTime: &stepTime,
		TimestampOrder: Order(condition.GetOrder()),
		Limit: int(condition.GetLimit()),
	}

	return &queryCondition
}
