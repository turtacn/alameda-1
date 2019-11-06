package v1alpha1

import (
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateNamespaces(ctx context.Context, in *ApiResources.CreateNamespacesRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNamespaces grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListNamespaces(ctx context.Context, in *ApiResources.ListNamespacesRequest) (*ApiResources.ListNamespacesResponse, error) {
	scope.Debug("Request received from ListNamespaces grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &ApiResources.ListNamespacesResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
	}, nil
}

func (s *ServiceV1alpha1) DeleteNamespaces(ctx context.Context, in *ApiResources.DeleteNamespacesRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteNamespaces grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}
