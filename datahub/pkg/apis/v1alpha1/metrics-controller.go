package v1alpha1

import (
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiMetrics "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/metrics"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateControllerMetrics(ctx context.Context, in *ApiMetrics.CreateControllerMetricsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateControllerMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListControllerMetrics(ctx context.Context, in *ApiMetrics.ListControllerMetricsRequest) (*ApiMetrics.ListControllerMetricsResponse, error) {
	scope.Debug("Request received from ListControllerMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &ApiMetrics.ListControllerMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
	}, nil
}
