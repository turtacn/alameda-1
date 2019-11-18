package v1alpha1

import (
	DaoPlannings "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/plannings"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateClusterPlannings(ctx context.Context, in *ApiPlannings.CreateClusterPlanningsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateClusterPlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	clusterDAO := DaoPlannings.NewClusterPlanningsDAO(*s.Config)
	err := clusterDAO.CreatePlannings(in)

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

func (s *ServiceV1alpha1) ListClusterPlannings(ctx context.Context, in *ApiPlannings.ListClusterPlanningsRequest) (*ApiPlannings.ListClusterPlanningsResponse, error) {
	scope.Debug("Request received from ListClusterPlannings grpc function: " + AlamedaUtils.InterfaceToString(in))

	clusterDAO := DaoPlannings.NewClusterPlanningsDAO(*s.Config)
	clusterPlannings, err := clusterDAO.ListPlannings(in)
	if err != nil {
		scope.Errorf("api ListClusterPlannings failed: %v", err)
		response := &ApiPlannings.ListClusterPlanningsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
			ClusterPlannings: clusterPlannings,
		}
		return response, nil
	}

	response := &ApiPlannings.ListClusterPlanningsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		ClusterPlannings: clusterPlannings,
	}

	return response, nil
}
