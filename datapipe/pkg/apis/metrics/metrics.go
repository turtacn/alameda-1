package metrics

import (
	DatapipeConfig "github.com/containers-ai/alameda/datapipe/pkg/config"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	Metrics "github.com/containers-ai/api/datapipe/metrics"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

var (
	scope = Log.RegisterScope("datapipe", "datapipe log", 0)
)

type ServiceMetric struct {
	Config *DatapipeConfig.Config
}

func NewServiceMetric(cfg *DatapipeConfig.Config) *ServiceMetric {
	service := ServiceMetric{}
	service.Config = cfg
	return &service
}

func (c *ServiceMetric) CreatePodMetrics(ctx context.Context, in *Metrics.CreatePodMetricsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePodMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (c *ServiceMetric) CreateNodeMetrics(ctx context.Context, in *Metrics.CreateNodeMetricsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNodeMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (c *ServiceMetric) ListPodMetrics(ctx context.Context, in *Metrics.ListPodMetricsRequest) (*Metrics.ListPodMetricsResponse, error) {
	scope.Debug("Request received from ListPodMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Metrics.ListPodMetricsResponse)
	return out, nil
}

func (c *ServiceMetric) ListNodeMetrics(ctx context.Context, in *Metrics.ListNodeMetricsRequest) (*Metrics.ListNodeMetricsResponse, error) {
	scope.Debug("Request received from ListPodMetrics grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Metrics.ListNodeMetricsResponse)
	return out, nil
}
