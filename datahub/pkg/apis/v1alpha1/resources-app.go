package v1alpha1

import (
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateApplications(ctx context.Context, in *ApiResources.CreateApplicationsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateApplications grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListApplications(ctx context.Context, in *ApiResources.ListApplicationsRequest) (*ApiResources.ListApplicationsResponse, error) {
	scope.Debug("Request received from ListApplications grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &ApiResources.ListApplicationsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
	}, nil
}
