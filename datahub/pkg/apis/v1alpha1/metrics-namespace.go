package v1alpha1

import (
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiMetrics "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/metrics"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateNamespaceMetrics(ctx context.Context, in *ApiMetrics.CreateNamespaceMetricsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNamespaceMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListNamespaceMetrics(ctx context.Context, in *ApiMetrics.ListNamespaceMetricsRequest) (*ApiMetrics.ListNamespaceMetricsResponse, error) {
	scope.Debug("Request received from ListNamespaceMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &ApiMetrics.ListNamespaceMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
	}, nil
}
