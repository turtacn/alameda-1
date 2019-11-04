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

/*
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
*/

// ListAlamedaPods returns predicted pods
func (s *ServiceV1alpha1) ListPods(ctx context.Context, in *ApiResources.ListPodsRequest) (*ApiResources.ListPodsResponse, error) {
	scope.Debug("Request received from ListAlamedaPods grpc function: " + AlamedaUtils.InterfaceToString(in))

	var containerDAO DaoClusterStatus.ContainerOperation = &DaoClusterStatusInflux.Container{
		InfluxDBConfig: *s.Config.InfluxDB,
	}

	namespace, name := "", ""
	if objectMeta := in.GetObjectMeta(); objectMeta != nil {
		namespace = objectMeta[0].GetNamespace()
		name = objectMeta[0].GetName()
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
