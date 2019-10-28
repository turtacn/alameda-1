package v1alpha1

import (
	DaoClusterStatus "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus"
	DaoClusterStatusInflux "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/influxdb"
	AlamedaUtils "github.com/containers-ai/alameda/pkg/utils"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"golang.org/x/net/context"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/genproto/googleapis/rpc/status"
)

// CreateAlamedaNodes add node information to database
func (s *ServiceV1alpha1) CreateAlamedaNodes(ctx context.Context, in *ApiResources.CreateAlamedaNodesRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateAlamedaNodes grpc function: " + AlamedaUtils.InterfaceToString(in))

	var nodeDAO DaoClusterStatus.NodeOperation = &DaoClusterStatusInflux.Node{
		InfluxDBConfig: *s.Config.InfluxDB,
	}
	if err := nodeDAO.RegisterAlamedaNodes(in.GetAlamedaNodes()); err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

// CreatePods add containers information of pods to database
func (s *ServiceV1alpha1) CreatePods(ctx context.Context, in *ApiResources.CreatePodsRequest) (*status.Status, error) {
	scope.Debug("Request received from CreatePods grpc function: " + AlamedaUtils.InterfaceToString(in))

	var containerDAO DaoClusterStatus.ContainerOperation = &DaoClusterStatusInflux.Container{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	if err := containerDAO.AddPods(in.GetPods()); err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) CreateControllers(ctx context.Context, in *ApiResources.CreateControllersRequest) (*status.Status, error) {
	scope.Debug("Request received from CreateControllers grpc function: " + AlamedaUtils.InterfaceToString(in))

	controllerDAO := &DaoClusterStatusInflux.Controller{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	err := controllerDAO.CreateControllers(in.GetControllers())
	if err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

// ListAlamedaPods returns predicted pods
func (s *ServiceV1alpha1) ListAlamedaPods(ctx context.Context, in *ApiResources.ListAlamedaPodsRequest) (*ApiResources.ListPodsResponse, error) {
	scope.Debug("Request received from ListAlamedaPods grpc function: " + AlamedaUtils.InterfaceToString(in))

	var containerDAO DaoClusterStatus.ContainerOperation = &DaoClusterStatusInflux.Container{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	namespace, name := "", ""
	if namespacedName := in.GetNamespacedName(); namespacedName != nil {
		namespace = namespacedName.GetNamespace()
		name = namespacedName.GetName()
	}
	kind := in.GetKind()
	timeRange := in.GetTimeRange()

	if alamedaPods, err := containerDAO.ListAlamedaPods(namespace, name, kind, timeRange); err != nil {
		scope.Error(err.Error())
		return &ApiResources.ListPodsResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	} else {
		res := &ApiResources.ListPodsResponse{
			Pods: alamedaPods,
			Status: &status.Status{
				Code: int32(code.Code_OK),
			},
		}
		scope.Debug("Request sent from ListAlamedaPods grpc function: " + AlamedaUtils.InterfaceToString(res))
		return res, nil
	}
}

// ListAlamedaNodes list nodes in cluster
func (s *ServiceV1alpha1) ListAlamedaNodes(ctx context.Context, in *ApiResources.ListAlamedaNodesRequest) (*ApiResources.ListNodesResponse, error) {
	scope.Debug("Request received from ListAlamedaNodes grpc function: " + AlamedaUtils.InterfaceToString(in))

	var nodeDAO DaoClusterStatus.NodeOperation = &DaoClusterStatusInflux.Node{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	timeRange := in.GetTimeRange()

	if alamedaNodes, err := nodeDAO.ListAlamedaNodes(timeRange); err != nil {
		scope.Error(err.Error())
		return &ApiResources.ListNodesResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	} else {
		return &ApiResources.ListNodesResponse{
			Status: &status.Status{
				Code: int32(code.Code_OK),
			},
			Nodes: alamedaNodes,
		}, nil
	}
}

func (s *ServiceV1alpha1) ListNodes(ctx context.Context, in *ApiResources.ListNodesRequest) (*ApiResources.ListNodesResponse, error) {
	scope.Debug("Request received from ListNodes grpc function: " + AlamedaUtils.InterfaceToString(in))

	var nodeDAO DaoClusterStatus.NodeOperation = &DaoClusterStatusInflux.Node{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	req := DaoClusterStatus.ListNodesRequest{
		NodeNames: in.GetNodeNames(),
		InCluster: true,
	}
	if nodes, err := nodeDAO.ListNodes(req); err != nil {
		scope.Error(err.Error())
		return &ApiResources.ListNodesResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	} else {
		return &ApiResources.ListNodesResponse{
			Status: &status.Status{
				Code: int32(code.Code_OK),
			},
			Nodes: nodes,
		}, nil
	}
}

func (s *ServiceV1alpha1) ListControllers(ctx context.Context, in *ApiResources.ListControllersRequest) (*ApiResources.ListControllersResponse, error) {
	scope.Debug("Request received from ListControllers grpc function: " + AlamedaUtils.InterfaceToString(in))

	controllerDAO := &DaoClusterStatusInflux.Controller{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	controllers, err := controllerDAO.ListControllers(in)
	if err != nil {
		scope.Error(err.Error())
		return &ApiResources.ListControllersResponse{
			Status: &status.Status{
				Code:    int32(code.Code_INTERNAL),
				Message: err.Error(),
			},
		}, nil
	}

	response := ApiResources.ListControllersResponse{
		Status: &status.Status{
			Code: int32(code.Code_OK),
		},
		Controllers: controllers,
	}
	return &response, nil
}

// ListPodsByNodeName list pods running on specific nodes
func (s *ServiceV1alpha1) ListPodsByNodeName(ctx context.Context, in *ApiResources.ListPodsByNodeNamesRequest) (*ApiResources.ListPodsResponse, error) {
	scope.Debug("Request received from ListPodsByNodeName grpc function: " + AlamedaUtils.InterfaceToString(in))

	return &ApiResources.ListPodsResponse{
		Status: &status.Status{
			Code:    int32(code.Code_OK),
			Message: "This function is deprecated.",
		},
	}, nil
}

// DeleteAlamedaNodes remove node information to database
func (s *ServiceV1alpha1) DeleteAlamedaNodes(ctx context.Context, in *ApiResources.DeleteAlamedaNodesRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteAlamedaNodes grpc function: " + AlamedaUtils.InterfaceToString(in))

	var nodeDAO DaoClusterStatus.NodeOperation = &DaoClusterStatusInflux.Node{
		InfluxDBConfig: *s.Config.InfluxDB,
	}
	alamedaNodeList := []*ApiResources.Node{}
	for _, alamedaNode := range in.GetAlamedaNodes() {
		alamedaNodeList = append(alamedaNodeList, &ApiResources.Node{
			Name: alamedaNode.GetName(),
		})
	}
	if err := nodeDAO.DeregisterAlamedaNodes(alamedaNodeList); err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

func (s *ServiceV1alpha1) DeleteControllers(ctx context.Context, in *ApiResources.DeleteControllersRequest) (*status.Status, error) {
	scope.Debug("Request received from DeleteControllers grpc function: " + AlamedaUtils.InterfaceToString(in))

	controllerDAO := &DaoClusterStatusInflux.Controller{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	err := controllerDAO.DeleteControllers(in)
	if err != nil {
		scope.Error(err.Error())
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}

	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}

// DeletePods update containers information of pods to database
func (s *ServiceV1alpha1) DeletePods(ctx context.Context, in *ApiResources.DeletePodsRequest) (*status.Status, error) {
	scope.Debug("Request received from DeletePods grpc function: " + AlamedaUtils.InterfaceToString(in))

	var containerDAO DaoClusterStatus.ContainerOperation = &DaoClusterStatusInflux.Container{
		InfluxDBConfig: *s.Config.InfluxDB,
	}
	if err := containerDAO.DeletePods(in.GetPods()); err != nil {
		scope.Errorf("DeletePods failed: %+v", err)
		return &status.Status{
			Code:    int32(code.Code_INTERNAL),
			Message: err.Error(),
		}, nil
	}
	return &status.Status{
		Code: int32(code.Code_OK),
	}, nil
}
