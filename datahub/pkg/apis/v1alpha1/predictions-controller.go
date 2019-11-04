package v1alpha1

import (
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateControllerPredictions(ctx context.Context, in *ApiPredictions.CreateControllerPredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateControllerPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListControllerPredictions(ctx context.Context, in *ApiPredictions.ListControllerPredictionsRequest) (*ApiPredictions.ListControllerPredictionsResponse, error) {
	scope.Debug("Request received from ListControllerPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &ApiPredictions.ListControllerPredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
	}, nil
}
