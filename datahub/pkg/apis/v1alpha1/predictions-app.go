package v1alpha1

import (
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateApplicationPredictions(ctx context.Context, in *ApiPredictions.CreateApplicationPredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateApplicationPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListApplicationPredictions(ctx context.Context, in *ApiPredictions.ListApplicationPredictionsRequest) (*ApiPredictions.ListApplicationPredictionsResponse, error) {
	scope.Debug("Request received from ListApplicationPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &ApiPredictions.ListApplicationPredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
	}, nil
}
