package v1alpha1

import (
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateClusterPredictions(ctx context.Context, in *ApiPredictions.CreateClusterPredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateClusterPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListClusterPredictions(ctx context.Context, in *ApiPredictions.ListClusterPredictionsRequest) (*ApiPredictions.ListClusterPredictionsResponse, error) {
	scope.Debug("Request received from ListClusterPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &ApiPredictions.ListClusterPredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
	}, nil
}
