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

func (s *ServiceV1alpha1) CreateNamespacePredictions(ctx context.Context, in *ApiPredictions.CreateNamespacePredictionsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNamespacePredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExtended := FormatRequest.CreateNamespacePredictionsRequestExtended{CreateNamespacePredictionsRequest: *in}
	if requestExtended.Validate() != nil {
		return &status.Status{
			Code: int32(code.Code_INVALID_ARGUMENT),
		}, nil
	}

	predictionDAO := DaoPrediction.NewNamespacePredictionsDAO(*s.Config)
	err := predictionDAO.CreatePredictions(requestExtended.ProducePredictions())
	if err != nil {
		scope.Errorf("failed to create namesapce predictions: %+v", err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) ListNamespacePredictions(ctx context.Context, in *ApiPredictions.ListNamespacePredictionsRequest) (*ApiPredictions.ListNamespacePredictionsResponse, error) {
	scope.Debug("Request received from ListNamespacePredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	requestExt := FormatRequest.ListNamespacePredictionsRequestExtended{Request: in}
	if err := requestExt.Validate(); err != nil {
		return &ApiPredictions.ListNamespacePredictionsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INVALID_ARGUMENT),
				Message: err.Error(),
			},
		}, nil
	}

	predictionDAO := DaoPrediction.NewNamespacePredictionsDAO(*s.Config)
	namespacesPredictionMap, err := predictionDAO.ListPredictions(requestExt.ProduceRequest())
	if err != nil {
		scope.Errorf("ListNamespacePredictions failed: %+v", err)
		return &ApiPredictions.ListNamespacePredictionsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	datahubNamespacePredictions := make([]*ApiPredictions.NamespacePrediction, 0)
	for _, namespacePrediction := range namespacesPredictionMap.MetricMap {
		namespacePredictionExtended := FormatResponse.NamespacePredictionExtended{NamespacePrediction: namespacePrediction}
		datahubNamespacePrediction := namespacePredictionExtended.ProducePredictions()
		datahubNamespacePredictions = append(datahubNamespacePredictions, datahubNamespacePrediction)
	}

	return &ApiPredictions.ListNamespacePredictionsResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		NamespacePredictions: datahubNamespacePredictions,
	}, nil
}
