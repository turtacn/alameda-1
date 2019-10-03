package requests

import (
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

var (
	scope = Log.RegisterScope("request-extend", "datahub(request-extend) log", 0)

	datahubAggregateFunction_DAOAggregateFunction = map[DatahubV1alpha1.TimeRange_AggregateFunction]DBCommon.AggregateFunction{
		DatahubV1alpha1.TimeRange_NONE: DBCommon.None,
		DatahubV1alpha1.TimeRange_MAX:  DBCommon.MaxOverTime,
	}
)
