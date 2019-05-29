package recommendations

import (
	DatapipeConfig "github.com/containers-ai/alameda/datapipe/pkg/config"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	Recommendations "github.com/containers-ai/api/datapipe/recommendations"
	"golang.org/x/net/context"
)

var (
	scope = Log.RegisterScope("datapipe", "datapipe log", 0)
)

type ServiceRecommendation struct {
	Config *DatapipeConfig.Config
}

func NewServiceRecommendation(cfg *DatapipeConfig.Config) *ServiceRecommendation {
	service := ServiceRecommendation{}
	service.Config = cfg
	return &service
}

func (c *ServiceRecommendation) ListPodRecommendations(ctx context.Context, in *Recommendations.ListPodRecommendationsRequest) (*Recommendations.ListPodRecommendationsResponse, error) {
	scope.Debug("Request received from ListPodRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Recommendations.ListPodRecommendationsResponse)
	return out, nil
}

func (c *ServiceRecommendation) ListAvailablePodRecommendations(ctx context.Context, in *Recommendations.ListPodRecommendationsRequest) (*Recommendations.ListPodRecommendationsResponse, error) {
	scope.Debug("Request received from ListAvailablePodRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Recommendations.ListPodRecommendationsResponse)
	return out, nil
}

func (c *ServiceRecommendation) ListControllerRecommendations(ctx context.Context, in *Recommendations.ListControllerRecommendationsRequest) (*Recommendations.ListControllerRecommendationsResponse, error) {
	scope.Debug("Request received from ListControllerRecommendations grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Recommendations.ListControllerRecommendationsResponse)
	return out, nil
}
