package agents

import (
	APIServerConfig "github.com/containers-ai/alameda/apiserver/pkg/config"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	Agents "github.com/containers-ai/federatorai-api/apiserver/agents"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

var (
	scope = Log.RegisterScope("apiserver", "apiserver log", 0)
)

type ServiceAgent struct {
	Config *APIServerConfig.Config
}

func NewServiceAgent(cfg *APIServerConfig.Config) *ServiceAgent {
	service := ServiceAgent{}
	service.Config = cfg
	return &service
}

func (c *ServiceAgent) Register(ctx context.Context, in *Agents.RegisterRequest) (*Agents.RegisterResponse, error) {
	scope.Debug("Request received from Register grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Agents.RegisterResponse)
	return out, nil
}

func (c *ServiceAgent) GetAgentVersion(ctx context.Context, in *Agents.GetAgentVersionRequest) (*Agents.GetAgentVersionResponse, error) {
	scope.Debug("Request received from GetAgentVersion grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Agents.GetAgentVersionResponse)
	return out, nil
}

func (c *ServiceAgent) DownloadAgent(ctx context.Context, in *Agents.DownloadAgentRequest) (*status.Status, error) {
	scope.Debug("Request received from DownloadAgent grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{Code: int32(code.Code_OK)}, nil
}
