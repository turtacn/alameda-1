package scores

import (
	DatapipeConfig "github.com/containers-ai/alameda/datapipe/pkg/config"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	Scores "github.com/containers-ai/api/datapipe/scores"
	"golang.org/x/net/context"
)

var (
	scope = Log.RegisterScope("datapipe", "datapipe log", 0)
)

type ServiceScore struct {
	Config *DatapipeConfig.Config
}

func NewServiceScore(cfg *DatapipeConfig.Config) *ServiceScore {
	service := ServiceScore{}
	service.Config = cfg
	return &service
}

func (c *ServiceScore) ListSimulatedSchedulingScores(ctx context.Context, in *Scores.ListSimulatedSchedulingScoresRequest) (*Scores.ListSimulatedSchedulingScoresResponse, error) {
	scope.Debug("Request received from ListSimulatedSchedulingScores grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Scores.ListSimulatedSchedulingScoresResponse)
	return out, nil
}
