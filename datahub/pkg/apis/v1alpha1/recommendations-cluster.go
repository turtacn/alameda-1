package v1alpha1

import (
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiRecommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateClusterRecommendations(ctx context.Context, in *ApiRecommendations.CreateClusterRecommendationsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateClusterRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListClusterRecommendations(ctx context.Context, in *ApiRecommendations.ListClusterRecommendationsRequest) (*ApiRecommendations.ListClusterRecommendationsResponse, error) {
	scope.Debug("Request received from ListClusterRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &ApiRecommendations.ListClusterRecommendationsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
	}, nil
}
