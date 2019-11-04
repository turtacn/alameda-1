package v1alpha1

import (
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateNamespacePlannings(ctx context.Context, in *ApiPlannings.CreateNamespacePlanningsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNamespacePlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListNamespacePlannings(ctx context.Context, in *ApiPlannings.ListNamespacePlanningsRequest) (*ApiPlannings.ListNamespacePlanningsResponse, error) {
	scope.Debug("Request received from ListNamespacePlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &ApiPlannings.ListNamespacePlanningsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
	}, nil
}
