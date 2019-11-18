package v1alpha1

import (
	DaoPlanning "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/plannings"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

// CreateControllerPlannings add controller plannings information to database
func (s *ServiceV1alpha1) CreateControllerPlannings(ctx context.Context, in *ApiPlannings.CreateControllerPlanningsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateControllerPlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	controllerDAO := DaoPlanning.NewControllerPlanningsDAO(*s.Config)
	err := controllerDAO.AddControllerPlannings(in)

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

// ListControllerPlannings list controller plannings
func (s *ServiceV1alpha1) ListControllerPlannings(ctx context.Context, in *ApiPlannings.ListControllerPlanningsRequest) (*ApiPlannings.ListControllerPlanningsResponse, error) {
	scope.Debug("Request received from ListControllerPlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	controllerDAO := DaoPlanning.NewControllerPlanningsDAO(*s.Config)
	controllerPlannings, err := controllerDAO.ListControllerPlannings(in)
	if err != nil {
		scope.Errorf("api ListControllerPlannings failed: %v", err)
		response := &ApiPlannings.ListControllerPlanningsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
			ControllerPlannings: controllerPlannings,
		}
		return response, nil
	}

	response := &ApiPlannings.ListControllerPlanningsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		ControllerPlannings: controllerPlannings,
	}

	scope.Debug("Response sent from ListControllerPlannings grpc function: " + AlamedaUtils.InterfaceToString(response))
	return response, nil
}
