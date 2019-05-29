package ping

import (
	APIServerConfig "github.com/containers-ai/alameda/apiserver/pkg/config"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	Ping "github.com/containers-ai/federatorai-api/apiserver/ping"
	"golang.org/x/net/context"
)

var (
	scope = Log.RegisterScope("apiserver", "apiserver log", 0)
)

type ServicePing struct {
	Config *APIServerConfig.Config
}

func NewServicePing(cfg *APIServerConfig.Config) *ServicePing {
	service := ServicePing{}
	service.Config = cfg
	return &service
}

func (c *ServicePing) Ping(ctx context.Context, in *Ping.PingRequest) (*Ping.PingResponse, error) {
	scope.Debug("Request received from Ping grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Ping.PingResponse)
	return out, nil
}
