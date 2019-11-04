package v1alpha1

import (
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateNamespacePredictions(ctx context.Context, in *ApiPredictions.CreateNamespacePredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNamespacePredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListNamespacePredictions(ctx context.Context, in *ApiPredictions.ListNamespacePredictionsRequest) (*ApiPredictions.ListNamespacePredictionsResponse, error) {
	scope.Debug("Request received from ListNamespacePredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &ApiPredictions.ListNamespacePredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
	}, nil
}
