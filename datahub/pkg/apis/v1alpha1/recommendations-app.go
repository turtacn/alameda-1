package v1alpha1

import (
	DaoRecommendation "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/recommendations"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiRecommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateApplicationRecommendations(ctx context.Context, in *ApiRecommendations.CreateApplicationRecommendationsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateApplicationRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	appRecommendationList := in.GetApplicationRecommendations()
	appDAO := DaoRecommendation.NewAppRecommendationsDAO(*s.Config)
	err := appDAO.CreateRecommendations(appRecommendationList)

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

func (s *ServiceV1alpha1) ListApplicationRecommendations(ctx context.Context, in *ApiRecommendations.ListApplicationRecommendationsRequest) (*ApiRecommendations.ListApplicationRecommendationsResponse, error) {
	scope.Debug("Request received from ListApplicationRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	appDAO := DaoRecommendation.NewAppRecommendationsDAO(*s.Config)
	appRecommendations, err := appDAO.ListRecommendations(in)
	if err != nil {
		scope.Errorf("api ListApplicationRecommendations failed: %v", err)
		response := &ApiRecommendations.ListApplicationRecommendationsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
			ApplicationRecommendations: appRecommendations,
		}
		return response, nil
	}

	response := &ApiRecommendations.ListApplicationRecommendationsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		ApplicationRecommendations: appRecommendations,
	}

	return response, nil
}
