package v1alpha1

import (
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateApplicationPlannings(ctx context.Context, in *ApiPlannings.CreateApplicationPlanningsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateApplicationPlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListApplicationPlannings(ctx context.Context, in *ApiPlannings.ListApplicationPlanningsRequest) (*ApiPlannings.ListApplicationPlanningsResponse, error) {
	scope.Debug("Request received from ListApplicationPlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &ApiPlannings.ListApplicationPlanningsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
	}, nil
}
