package v1alpha1

import (
	DaoRecommendation "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/recommendations"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiRecommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateNamespaceRecommendations(ctx context.Context, in *ApiRecommendations.CreateNamespaceRecommendationsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNamespaceRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	namespaceRecommendationList := in.GetNamespaceRecommendations()
	namespaceDAO := DaoRecommendation.NewNamespaceRecommendationsDAO(*s.Config)
	err := namespaceDAO.CreateRecommendations(namespaceRecommendationList)

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

func (s *ServiceV1alpha1) ListNamespaceRecommendations(ctx context.Context, in *ApiRecommendations.ListNamespaceRecommendationsRequest) (*ApiRecommendations.ListNamespaceRecommendationsResponse, error) {
	scope.Debug("Request received from ListNamespaceRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	namespaceDAO := DaoRecommendation.NewNamespaceRecommendationsDAO(*s.Config)
	namespaceRecommendations, err := namespaceDAO.ListRecommendations(in)
	if err != nil {
		scope.Errorf("api ListNamespaceRecommendations failed: %v", err)
		response := &ApiRecommendations.ListNamespaceRecommendationsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
			NamespaceRecommendations: namespaceRecommendations,
		}
		return response, nil
	}

	response := &ApiRecommendations.ListNamespaceRecommendationsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		NamespaceRecommendations: namespaceRecommendations,
	}

	return response, nil
}
