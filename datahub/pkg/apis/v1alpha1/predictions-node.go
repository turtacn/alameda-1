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

// CreateNodePredictions add node predictions information to database
func (s *ServiceV1alpha1) CreateNodePredictions(ctx context.Context, in *ApiPredictions.CreateNodePredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNodePredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExtended := FormatRequest.CreateNodePredictionsRequestExtended{CreateNodePredictionsRequest: *in}
	if requestExtended.Validate() != nil {
		return &status.Status{
			Code: int32(code.Code_INVALID_ARGUMENT),
		}, nil
	}

	predictionDAO := DaoPrediction.NewNodePredictionsDAO(*s.Config)
	err := predictionDAO.CreatePredictions(requestExtended.ProducePredictions())
	if err != nil {
		scope.Errorf("failed to create node predictions: %+v", err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

// ListNodePredictions list nodes' predictions
func (s *ServiceV1alpha1) ListNodePredictions(ctx context.Context, in *ApiPredictions.ListNodePredictionsRequest) (*ApiPredictions.ListNodePredictionsResponse, error) {
	scope.Debug("Request received from ListNodePredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExt := FormatRequest.ListNodePredictionsRequestExtended{Request: in}
	if err := requestExt.Validate(); err != nil {
		return &ApiPredictions.ListNodePredictionsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}

	predictionDAO := DaoPrediction.NewNodePredictionsDAO(*s.Config)
	nodesPredictionMap, err := predictionDAO.ListPredictions(requestExt.ProduceRequest())
	if err != nil {
		scope.Errorf("ListNodePredictions failed: %+v", err)
		return &ApiPredictions.ListNodePredictionsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	datahubNodePredictions := make([]*ApiPredictions.NodePrediction, 0)
	for _, nodePrediction := range nodesPredictionMap.MetricMap {
		nodePredictionExtended := FormatResponse.NodePredictionExtended{NodePrediction: nodePrediction}
		datahubNodePrediction := nodePredictionExtended.ProducePredictions()
		datahubNodePredictions = append(datahubNodePredictions, datahubNodePrediction)
	}

	return &ApiPredictions.ListNodePredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		NodePredictions: datahubNodePredictions,
	}, nil
}
