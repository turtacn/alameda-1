package resources

import (
	DatapipeConfig "github.com/containers-ai/alameda/datapipe/pkg/config"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	Resources "github.com/containers-ai/api/datapipe/resources"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

var (
	scope = Log.RegisterScope("datapipe", "datapipe log", 0)
)

type ServiceResource struct {
	Config *DatapipeConfig.Config
}

func NewServiceResource(cfg *DatapipeConfig.Config) *ServiceResource {
	service := ServiceResource{}
	service.Config = cfg
	return &service
}

func (c *ServiceResource) CreateContainers(ctx context.Context, in *Resources.CreateContainersRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateContainers grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (c *ServiceResource) CreatePods(ctx context.Context, in *Resources.CreatePodsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePods grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (c *ServiceResource) CreateControllers(ctx context.Context, in *Resources.CreateControllersRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateControllers grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (c *ServiceResource) CreateNodes(ctx context.Context, in *Resources.CreateNodesRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateNodes grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (c *ServiceResource) ListContainers(ctx context.Context, in *Resources.ListContainersRequest) (*Resources.ListContainersResponse, error) {
	scope.Debug("Request received from ListContainers grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Resources.ListContainersResponse)
	return out, nil
}

func (c *ServiceResource) ListPods(ctx context.Context, in *Resources.ListPodsRequest) (*Resources.ListPodsResponse, error) {
	scope.Debug("Request received from ListPods grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Resources.ListPodsResponse)
	return out, nil
}

func (c *ServiceResource) ListPodsByNodeName(ctx context.Context, in *Resources.ListPodsByNodeNamesRequest) (*Resources.ListPodsResponse, error) {
	scope.Debug("Request received from ListPodsByNodeName grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Resources.ListPodsResponse)
	return out, nil
}

func (c *ServiceResource) ListControllers(ctx context.Context, in *Resources.ListControllersRequest) (*Resources.ListControllersResponse, error) {
	scope.Debug("Request received from ListControllers grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Resources.ListControllersResponse)
	return out, nil
}

func (c *ServiceResource) ListNodes(ctx context.Context, in *Resources.ListNodesRequest) (*Resources.ListNodesResponse, error) {
	scope.Debug("Request received from ListNodes grpc function: " + AlamedaUtils.InterfaceToString(in))

	out := new(Resources.ListNodesResponse)
	return out, nil
}

func (c *ServiceResource) DeleteContainers(ctx context.Context, in *Resources.DeleteContainersRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteContainers grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (c *ServiceResource) DeletePods(ctx context.Context, in *Resources.DeletePodsRequest) (*status.Status, error) {
	scope.Debug("Request received from DeletePods grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (c *ServiceResource) DeleteControllers(ctx context.Context, in *Resources.DeleteControllersRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteControllers grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{Code: int32(code.Code_OK)}, nil
}

func (c *ServiceResource) DeleteNodes(ctx context.Context, in *Resources.DeleteNodesRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteNodes grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &status.Status{Code: int32(code.Code_OK)}, nil
}
