package v1alpha1

import (
	DaoPlannings "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/plannings"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateNamespacePlannings(ctx context.Context, in *ApiPlannings.CreateNamespacePlanningsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNamespacePlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	namespaceDAO := DaoPlannings.NewNamespacePlanningsDAO(*s.Config)
	err := namespaceDAO.CreatePlannings(in)

	if err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, err
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListNamespacePlannings(ctx context.Context, in *ApiPlannings.ListNamespacePlanningsRequest) (*ApiPlannings.ListNamespacePlanningsResponse, error) {
	scope.Debug("Request received from ListNamespacePlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	namespaceDAO := DaoPlannings.NewNamespacePlanningsDAO(*s.Config)
	namespacePlannings, err := namespaceDAO.ListPlannings(in)
	if err != nil {
		scope.Errorf("api ListNamespacePlannings failed: %v", err)
		response := &ApiPlannings.ListNamespacePlanningsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
			NamespacePlannings: namespacePlannings,
		}
		return response, nil
	}

	response := &ApiPlannings.ListNamespacePlanningsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		NamespacePlannings: namespacePlannings,
	}

	return response, nil
}
