package v1alpha1

import (
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiMetrics "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/metrics"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateClusterMetrics(ctx context.Context, in *ApiMetrics.CreateClusterMetricsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateClusterMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListClusterMetrics(ctx context.Context, in *ApiMetrics.ListClusterMetricsRequest) (*ApiMetrics.ListClusterMetricsResponse, error) {
	scope.Debug("Request received from ListClusterMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &ApiMetrics.ListClusterMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
	}, nil
}
