package requests

import (
	DaoScore "github.com/containers-ai/alameda/datahub/pkg/dao/score"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

type ListSimulatedSchedulingScoresRequestExtended struct {
	Request *DatahubV1alpha1.ListSimulatedSchedulingScoresRequest
}

func (r *ListSimulatedSchedulingScoresRequestExtended) ProduceRequest() DaoScore.ListRequest {
	var (
		queryCondition DBCommon.QueryCondition
	)

	queryCondition = QueryConditionExtend{r.Request.GetQueryCondition()}.QueryCondition()
	listRequest := DaoScore.ListRequest{
		QueryCondition: queryCondition,
	}

	return listRequest
}
