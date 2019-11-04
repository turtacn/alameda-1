package v1alpha1

import (
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiMetrics "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/metrics"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateApplicationMetrics(ctx context.Context, in *ApiMetrics.CreateApplicationMetricsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateApplicationMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListApplicationMetrics(ctx context.Context, in *ApiMetrics.ListApplicationMetricsRequest) (*ApiMetrics.ListApplicationMetricsResponse, error) {
	scope.Debug("Request received from ListApplicationMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &ApiMetrics.ListApplicationMetricsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
	}, nil
}
