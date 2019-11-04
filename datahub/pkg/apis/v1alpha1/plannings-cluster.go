package v1alpha1

import (
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateClusterPlannings(ctx context.Context, in *ApiPlannings.CreateClusterPlanningsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateClusterPlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListClusterPlannings(ctx context.Context, in *ApiPlannings.ListClusterPlanningsRequest) (*ApiPlannings.ListClusterPlanningsResponse, error) {
	scope.Debug("Request received from ListClusterPlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &ApiPlannings.ListClusterPlanningsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
	}, nil
}
