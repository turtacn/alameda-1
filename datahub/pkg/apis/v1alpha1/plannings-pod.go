package v1alpha1

import (
	DaoPlanning "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/plannings"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

// CreatePodPlannings add pod plannings information to database
func (s *ServiceV1alpha1) CreatePodPlannings(ctx context.Context, in *ApiPlannings.CreatePodPlanningsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePodPlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	containerDAO := DaoPlanning.NewContainerPlanningsDAO(*s.Config)
	if err := containerDAO.AddPodPlannings(in); err != nil {
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

// ListPodPlannings list pod plannings
func (s *ServiceV1alpha1) ListPodPlannings(ctx context.Context, in *ApiPlannings.ListPodPlanningsRequest) (*ApiPlannings.ListPodPlanningsResponse, error) {
	scope.Debug("Request received from ListPodPlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	containerDAO := DaoPlanning.NewContainerPlanningsDAO(*s.Config)
	podPlannings, err := containerDAO.ListPodPlannings(in)
	if err != nil {
		scope.Error(err.Error())
		return &ApiPlannings.ListPodPlanningsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	res := &ApiPlannings.ListPodPlanningsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		PodPlannings: podPlannings,
	}
	scope.Debug("Response sent from ListPodPlannings grpc function: " + AlamedaUtils.InterfaceToString(res))
	return res, nil
}
