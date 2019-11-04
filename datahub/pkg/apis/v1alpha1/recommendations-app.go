package v1alpha1

import (
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiRecommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateApplicationRecommendations(ctx context.Context, in *ApiRecommendations.CreateApplicationRecommendationsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateApplicationRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListApplicationRecommendations(ctx context.Context, in *ApiRecommendations.ListApplicationRecommendationsRequest) (*ApiRecommendations.ListApplicationRecommendationsResponse, error) {
	scope.Debug("Request received from ListApplicationRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &ApiRecommendations.ListApplicationRecommendationsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
	}, nil
}
