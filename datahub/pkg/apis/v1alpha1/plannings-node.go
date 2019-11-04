package v1alpha1

import (
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateNodePlannings(ctx context.Context, in *ApiPlannings.CreateNodePlanningsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNodePlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListNodePlannings(ctx context.Context, in *ApiPlannings.ListNodePlanningsRequest) (*ApiPlannings.ListNodePlanningsResponse, error) {
	scope.Debug("Request received from ListNodePlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &ApiPlannings.ListNodePlanningsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
	}, nil
}
