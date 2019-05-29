package ping

import (
	DatapipeConfig "github.com/containers-ai/alameda/datapipe/pkg/config"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	Ping "github.com/containers-ai/api/datapipe/ping"
	"golang.org/x/net/context"
)

var (
	scope = Log.RegisterScope("datapipe", "datapipe log", 0)
)

type ServicePing struct {
	Config *DatapipeConfig.Config
}

func NewServicePing(cfg *DatapipeConfig.Config) *ServicePing {
	service := ServicePing{}
	service.Config = cfg
	return &service
}

func (c *ServicePing) Ping(ctx context.Context, in *Ping.PingRequest) (*Ping.PingResponse, error) {
	scope.Debug("Request received from Ping grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Ping.PingResponse)
	return out, nil
}
