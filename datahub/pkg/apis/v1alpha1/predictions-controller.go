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

func (s *ServiceV1alpha1) CreateControllerPredictions(ctx context.Context, in *ApiPredictions.CreateControllerPredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateControllerPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExtended := FormatRequest.CreateControllerPredictionsRequestExtended{CreateControllerPredictionsRequest: *in}
	if requestExtended.Validate() != nil {
		return &status.Status{
			Code: int32(code.Code_INVALID_ARGUMENT),
		}, nil
	}

	predictionDAO := DaoPrediction.NewControllerPredictionsDAO(*s.Config)
	err := predictionDAO.CreatePredictions(requestExtended.ProducePredictions())
	if err != nil {
		scope.Errorf("failed to create controller predictions: %+v", err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListControllerPredictions(ctx context.Context, in *ApiPredictions.ListControllerPredictionsRequest) (*ApiPredictions.ListControllerPredictionsResponse, error) {
	scope.Debug("Request received from ListControllerPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExt := FormatRequest.ListControllerPredictionsRequestExtended{Request: in}
	if err := requestExt.Validate(); err != nil {
		return &ApiPredictions.ListControllerPredictionsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}

	predictionDAO := DaoPrediction.NewControllerPredictionsDAO(*s.Config)
	controllersPredictionMap, err := predictionDAO.ListPredictions(requestExt.ProduceRequest())
	if err != nil {
		scope.Errorf("ListControllerPredictions failed: %+v", err)
		return &ApiPredictions.ListControllerPredictionsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	datahubControllerPredictions := make([]*ApiPredictions.ControllerPrediction, 0)
	for _, controllerPrediction := range controllersPredictionMap.MetricMap {
		controllerPredictionExtended := FormatResponse.ControllerPredictionExtended{ControllerPrediction: controllerPrediction}
		datahubControllerPrediction := controllerPredictionExtended.ProducePredictions()
		datahubControllerPredictions = append(datahubControllerPredictions, datahubControllerPrediction)
	}

	return &ApiPredictions.ListControllerPredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		ControllerPredictions: datahubControllerPredictions,
	}, nil
}
