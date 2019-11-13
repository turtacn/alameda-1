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

func (s *ServiceV1alpha1) CreateApplicationPredictions(ctx context.Context, in *ApiPredictions.CreateApplicationPredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateApplicationPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExtended := FormatRequest.CreateApplicationPredictionsRequestExtended{CreateApplicationPredictionsRequest: *in}
	if requestExtended.Validate() != nil {
		return &status.Status{
			Code: int32(code.Code_INVALID_ARGUMENT),
		}, nil
	}

	predictionDAO := DaoPrediction.NewApplicationPredictionsDAO(*s.Config)
	err := predictionDAO.CreatePredictions(requestExtended.ProducePredictions())
	if err != nil {
		scope.Errorf("failed to create application predictions: %+v", err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListApplicationPredictions(ctx context.Context, in *ApiPredictions.ListApplicationPredictionsRequest) (*ApiPredictions.ListApplicationPredictionsResponse, error) {
	scope.Debug("Request received from ListApplicationPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExt := FormatRequest.ListApplicationPredictionsRequestExtended{Request: in}
	if err := requestExt.Validate(); err != nil {
		return &ApiPredictions.ListApplicationPredictionsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}

	predictionDAO := DaoPrediction.NewApplicationPredictionsDAO(*s.Config)
	applicationsPredictionMap, err := predictionDAO.ListPredictions(requestExt.ProduceRequest())
	if err != nil {
		scope.Errorf("ListApplicationPredictions failed: %+v", err)
		return &ApiPredictions.ListApplicationPredictionsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	datahubApplicationPredictions := make([]*ApiPredictions.ApplicationPrediction, 0)
	for _, applicationPrediction := range applicationsPredictionMap.MetricMap {
		applicationPredictionExtended := FormatResponse.ApplicationPredictionExtended{ApplicationPrediction: applicationPrediction}
		datahubApplicationPrediction := applicationPredictionExtended.ProducePredictions()
		datahubApplicationPredictions = append(datahubApplicationPredictions, datahubApplicationPrediction)
	}

	return &ApiPredictions.ListApplicationPredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		ApplicationPredictions: datahubApplicationPredictions,
	}, nil
}
