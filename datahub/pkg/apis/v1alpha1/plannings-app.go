package v1alpha1

import (
	DaoPlannings "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/plannings"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateApplicationPlannings(ctx context.Context, in *ApiPlannings.CreateApplicationPlanningsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateApplicationPlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	appDAO := DaoPlannings.NewAppPlanningsDAO(*s.Config)
	err := appDAO.CreatePlannings(in)

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

func (s *ServiceV1alpha1) ListApplicationPlannings(ctx context.Context, in *ApiPlannings.ListApplicationPlanningsRequest) (*ApiPlannings.ListApplicationPlanningsResponse, error) {
	scope.Debug("Request received from ListApplicationPlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	appDAO := DaoPlannings.NewAppPlanningsDAO(*s.Config)
	appRecommendations, err := appDAO.ListPlannings(in)
	if err != nil {
		scope.Errorf("api ListApplicationPlannings failed: %v", err)
		response := &ApiPlannings.ListApplicationPlanningsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
			ApplicationPlannings: appRecommendations,
		}
		return response, nil
	}

	response := &ApiPlannings.ListApplicationPlanningsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		ApplicationPlannings: appRecommendations,
	}

	return response, nil
}
