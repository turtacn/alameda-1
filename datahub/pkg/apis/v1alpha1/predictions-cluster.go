package v1alpha1

import (
	DaoPrediction "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions"
	FormatRequest "github.com/containers-ai/alameda/datahub/pkg/formatconversion/requests"
	FormatResponse "github.com/containers-ai/alameda/datahub/pkg/formatconversion/responses"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

func (s *ServiceV1alpha1) CreateClusterPredictions(ctx context.Context, in *ApiPredictions.CreateClusterPredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateClusterPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExtended := FormatRequest.CreateClusterPredictionsRequestExtended{CreateClusterPredictionsRequest: *in}
	if requestExtended.Validate() != nil {
		return &status.Status{
			Code: int32(code.Code_INVALID_ARGUMENT),
		}, nil
	}

	predictionDAO := DaoPrediction.NewClusterPredictionsDAO(*s.Config)
	err := predictionDAO.CreatePredictions(requestExtended.ProducePredictions())
	if err != nil {
		scope.Errorf("failed to create cluster predictions: %+v", err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListClusterPredictions(ctx context.Context, in *ApiPredictions.ListClusterPredictionsRequest) (*ApiPredictions.ListClusterPredictionsResponse, error) {
	scope.Debug("Request received from ListClusterPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExt := FormatRequest.ListClusterPredictionsRequestExtended{Request: in}
	if err := requestExt.Validate(); err != nil {
		return &ApiPredictions.ListClusterPredictionsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}

	predictionDAO := DaoPrediction.NewClusterPredictionsDAO(*s.Config)
	clustersPredictionMap, err := predictionDAO.ListPredictions(requestExt.ProduceRequest())
	if err != nil {
		scope.Errorf("ListClusterPredictions failed: %+v", err)
		return &ApiPredictions.ListClusterPredictionsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	datahubClusterPredictions := make([]*ApiPredictions.ClusterPrediction, 0)
	for _, clusterPrediction := range clustersPredictionMap.MetricMap {
		clusterPredictionExtended := FormatResponse.ClusterPredictionExtended{ClusterPrediction: clusterPrediction}
		datahubClusterPrediction := clusterPredictionExtended.ProducePredictions()
		datahubClusterPredictions = append(datahubClusterPredictions, datahubClusterPrediction)
	}

	return &ApiPredictions.ListClusterPredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		ClusterPredictions: datahubClusterPredictions,
	}, nil
}
