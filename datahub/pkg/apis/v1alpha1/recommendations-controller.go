package v1alpha1

import (
	DaoRecommendation "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/recommendations"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiRecommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

// CreatePodRecommendations add controller recommendations information to database
func (s *ServiceV1alpha1) CreateControllerRecommendations(ctx context.Context, in *ApiRecommendations.CreateControllerRecommendationsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateControllerRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	controllerRecommendationList := in.GetControllerRecommendations()
	controllerDAO := DaoRecommendation.NewControllerRecommendationsDAO(*s.Config)
	err := controllerDAO.AddControllerRecommendations(controllerRecommendationList)

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

// ListControllerRecommendations list controller recommendations
func (s *ServiceV1alpha1) ListControllerRecommendations(ctx context.Context, in *ApiRecommendations.ListControllerRecommendationsRequest) (*ApiRecommendations.ListControllerRecommendationsResponse, error) {
	scope.Debug("Request received from ListControllerRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	controllerDAO := DaoRecommendation.NewControllerRecommendationsDAO(*s.Config)
	controllerRecommendations, err := controllerDAO.ListControllerRecommendations(in)
	if err != nil {
		scope.Errorf("api ListControllerRecommendations failed: %v", err)
		response := &ApiRecommendations.ListControllerRecommendationsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
			ControllerRecommendations: controllerRecommendations,
		}
		return response, nil
	}

	response := &ApiRecommendations.ListControllerRecommendationsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		ControllerRecommendations: controllerRecommendations,
	}

	scope.Debug("Response sent from ListControllerRecommendations grpc function: " + AlamedaUtils.InterfaceToString(response))
	return response, nil
}
