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

func (s *ServiceV1alpha1) CreateNamespaces(ctx context.Context, in *ApiResources.CreateNamespacesRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNamespaces grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExtended := FormatRequest.CreateNamespacesRequestExtended{CreateNamespacesRequest: *in}
	if requestExtended.Validate() != nil {
		return &status.Status{
			Code: int32(code.Code_INVALID_ARGUMENT),
		}, nil
	}

	namespaceDAO := DaoCluster.NewNamespaceDAO(*s.Config)
	if err := namespaceDAO.CreateNamespaces(requestExtended.ProduceNamespaces()); err != nil {
		scope.Errorf("failed to create namespaces: %+v", err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListNamespaces(ctx context.Context, in *ApiResources.ListNamespacesRequest) (*ApiResources.ListNamespacesResponse, error) {
	scope.Debug("Request received from ListNamespaces grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExt := FormatRequest.ListNamespacesRequestExtended{ListNamespacesRequest: in}
	if err := requestExt.Validate(); err != nil {
		return &ApiResources.ListNamespacesResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}

	namespaceDAO := DaoCluster.NewNamespaceDAO(*s.Config)
	nss, err := namespaceDAO.ListNamespaces(requestExt.ProduceRequest())
	if err != nil {
		scope.Errorf("ListNamespaces failed: %+v", err)
		return &ApiResources.ListNamespacesResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	namespaces := make([]*ApiResources.Namespace, 0)
	for _, ns := range nss {
		namespaceExtended := FormatResponse.NamespaceExtended{Namespace: ns}
		namespace := namespaceExtended.ProduceNamespace()
		namespaces = append(namespaces, namespace)
	}

	return &ApiResources.ListNamespacesResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Namespaces: namespaces,
	}, nil
}

func (s *ServiceV1alpha1) DeleteNamespaces(ctx context.Context, in *ApiResources.DeleteNamespacesRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteNamespaces grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExt := FormatRequest.DeleteNamespacesRequestExtended{DeleteNamespacesRequest: in}
	if err := requestExt.Validate(); err != nil {
		return &status.Status{
			Code:    int32(code.Code_INVALID_ARGUMENT),
			Message: err.Error(),
		}, nil
	}

	namespaceDAO := DaoCluster.NewNamespaceDAO(*s.Config)
	if err := namespaceDAO.DeleteNamespaces(requestExt.ProduceRequest()); err != nil {
		scope.Errorf("failed to delete namespaces: %+v", err)
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}
