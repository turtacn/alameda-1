package predictions

import (
	DatapipeConfig "github.com/containers-ai/alameda/datapipe/pkg/config"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	Predictions "github.com/containers-ai/api/datapipe/predictions"
	"golang.org/x/net/context"
)

var (
	scope = Log.RegisterScope("datapipe", "datapipe log", 0)
)

type ServicePrediction struct {
	Config *DatapipeConfig.Config
}

func NewServicePrediction(cfg *DatapipeConfig.Config) *ServicePrediction {
	service := ServicePrediction{}
	service.Config = cfg
	return &service
}

func (c *ServicePrediction) ListPodPredictions(ctx context.Context, in *Predictions.ListPodPredictionsRequest) (*Predictions.ListPodPredictionsResponse, error) {
	scope.Debug("Request received from ListPodPredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Predictions.ListPodPredictionsResponse)
	return out, nil
}

func (c *ServicePrediction) ListNodePredictions(ctx context.Context, in *Predictions.ListNodePredictionsRequest) (*Predictions.ListNodePredictionsResponse, error) {
	scope.Debug("Request received from ListNodePredictions grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Predictions.ListNodePredictionsResponse)
	return out, nil
}
