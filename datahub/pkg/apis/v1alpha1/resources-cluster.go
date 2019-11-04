package v1alpha1

import (
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateClusters(ctx context.Context, in *ApiResources.CreateClustersRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateClusters grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListClusters(ctx context.Context, in *ApiResources.ListClustersRequest) (*ApiResources.ListClustersResponse, error) {
	scope.Debug("Request received from ListClusters grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &ApiResources.ListClustersResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
	}, nil
}
