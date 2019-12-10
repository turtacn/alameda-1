package influxdb

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/clusterstatus"
	Metadata "github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

type Controller struct {
	InfluxDBConfig InternalInflux.Config
}

func NewControllerWithConfig(config InternalInflux.Config) DaoClusterTypes.ControllerDAO {
	return &Controller{InfluxDBConfig: config}
}

func (p *Controller) CreateControllers(controllers []*DaoClusterTypes.Controller) error {
	controllerRepo := RepoInfluxCluster.NewControllerRepository(p.InfluxDBConfig)
	if err := controllerRepo.CreateControllers(controllers); err != nil {
		scope.Error(err.Error())
		return err
	}
	return nil
}

func (p *Controller) ListControllers(request *DaoClusterTypes.ListControllersRequest) ([]*DaoClusterTypes.Controller, error) {
	controllerRepo := RepoInfluxCluster.NewControllerRepository(p.InfluxDBConfig)
	controllers, err := controllerRepo.ListControllers(request)
	if err != nil {
		scope.Error(err.Error())
		return make([]*DaoClusterTypes.Controller, 0), err
	}
	return controllers, nil
}

func (p *Controller) DeleteControllers(request *DaoClusterTypes.DeleteControllersRequest) error {
	delPodsReq := p.genDeletePodsRequest(request)

	// Delete controllers
	controllerRepo := RepoInfluxCluster.NewControllerRepository(p.InfluxDBConfig)
	if err := controllerRepo.DeleteControllers(request); err != nil {
		scope.Error(err.Error())
		return err
	}

	// Delete pods
	podDAO := NewPodWithConfig(p.InfluxDBConfig)
	if err := podDAO.DeletePods(delPodsReq); err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}

func (p *Controller) genDeletePodsRequest(request *DaoClusterTypes.DeleteControllersRequest) *DaoClusterTypes.DeletePodsRequest {
	delPodsReq := DaoClusterTypes.NewDeletePodsRequest()

	for _, controllerObjectMeta := range request.ControllerObjectMeta {
		topController := &Metadata.ObjectMeta{}
		alamedaScaler := &Metadata.ObjectMeta{}

		if controllerObjectMeta.ObjectMeta != nil {
			topController.Name = controllerObjectMeta.ObjectMeta.Name
			topController.Namespace = controllerObjectMeta.ObjectMeta.Namespace
			topController.ClusterName = controllerObjectMeta.ObjectMeta.ClusterName
		} else {
			topController = nil
		}

		if controllerObjectMeta.AlamedaScaler != nil {
			alamedaScaler.Name = controllerObjectMeta.AlamedaScaler.Name
			alamedaScaler.Namespace = controllerObjectMeta.AlamedaScaler.Namespace
			alamedaScaler.ClusterName = controllerObjectMeta.AlamedaScaler.ClusterName
		} else {
			alamedaScaler = nil
		}

		podObjectMeta := DaoClusterTypes.NewPodObjectMeta(nil, topController, alamedaScaler, controllerObjectMeta.Kind, controllerObjectMeta.ScalingTool)
		delPodsReq.PodObjectMeta = append(delPodsReq.PodObjectMeta, podObjectMeta)
	}

	return delPodsReq
}
