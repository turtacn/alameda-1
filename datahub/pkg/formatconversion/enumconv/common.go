package enumconv

import (
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
)

var AggregateFunctionNameMap map[ApiCommon.TimeRange_AggregateFunction]DBCommon.AggregateFunction = map[ApiCommon.TimeRange_AggregateFunction]DBCommon.AggregateFunction{
	ApiCommon.TimeRange_NONE: DBCommon.None,
	ApiCommon.TimeRange_MAX:  DBCommon.MaxOverTime,
	ApiCommon.TimeRange_AVG:  DBCommon.AvgOverTime,
}
