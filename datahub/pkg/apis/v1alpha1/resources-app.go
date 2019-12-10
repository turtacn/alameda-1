package v1alpha1

import (
	DaoCluster "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus"
	FormatRequest "github.com/containers-ai/alameda/datahub/pkg/formatconversion/requests"
	FormatResponse "github.com/containers-ai/alameda/datahub/pkg/formatconversion/responses"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateApplications(ctx context.Context, in *ApiResources.CreateApplicationsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateApplications grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExtended := FormatRequest.CreateApplicationsRequestExtended{CreateApplicationsRequest: *in}
	if requestExtended.Validate() != nil {
		return &status.Status{
			Code: int32(code.Code_INVALID_ARGUMENT),
		}, nil
	}

	applicationDAO := DaoCluster.NewApplicationDAO(*s.Config)
	if err := applicationDAO.CreateApplications(requestExtended.ProduceApplications()); err != nil {
		scope.Errorf("failed to create applications: %+v", err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListApplications(ctx context.Context, in *ApiResources.ListApplicationsRequest) (*ApiResources.ListApplicationsResponse, error) {
	scope.Debug("Request received from ListApplications grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExt := FormatRequest.ListApplicationsRequestExtended{ListApplicationsRequest: in}
	if err := requestExt.Validate(); err != nil {
		return &ApiResources.ListApplicationsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}

	applicationDAO := DaoCluster.NewApplicationDAO(*s.Config)
	apps, err := applicationDAO.ListApplications(requestExt.ProduceRequest())
	if err != nil {
		scope.Errorf("ListApplications failed: %+v", err)
		return &ApiResources.ListApplicationsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	applications := make([]*ApiResources.Application, 0)
	for _, app := range apps {
		applicationExtended := FormatResponse.ApplicationExtended{Application: app}
		application := applicationExtended.ProduceApplication()
		applications = append(applications, application)
	}

	return &ApiResources.ListApplicationsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Applications: applications,
	}, nil
}

func (s *ServiceV1alpha1) DeleteApplications(ctx context.Context, in *ApiResources.DeleteApplicationsRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteApplications grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExt := FormatRequest.DeleteApplicationsRequestExtended{DeleteApplicationsRequest: in}
	if err := requestExt.Validate(); err != nil {
		return &status.Status{
			Code:    int32(code.Code_INVALID_ARGUMENT),
			Message: err.Error(),
		}, nil
	}

	applicationDAO := DaoCluster.NewApplicationDAO(*s.Config)
	if err := applicationDAO.DeleteApplications(requestExt.ProduceRequest()); err != nil {
		scope.Errorf("failed to delete applications: %+v", err)
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}
