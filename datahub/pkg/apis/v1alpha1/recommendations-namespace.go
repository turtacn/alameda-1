package v1alpha1

import (
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiRecommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateNamespaceRecommendations(ctx context.Context, in *ApiRecommendations.CreateNamespaceRecommendationsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNamespaceRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListNamespaceRecommendations(ctx context.Context, in *ApiRecommendations.ListNamespaceRecommendationsRequest) (*ApiRecommendations.ListNamespaceRecommendationsResponse, error) {
	scope.Debug("Request received from ListNamespaceRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &ApiRecommendations.ListNamespaceRecommendationsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
	}, nil
}
