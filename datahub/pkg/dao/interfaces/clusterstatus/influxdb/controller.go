package influxdb

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/clusterstatus"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type Controller struct {
	InfluxDBConfig InternalInflux.Config
}

func NewControllerWithConfig(config InternalInflux.Config) DaoClusterTypes.ControllerDAO {
	return &Controller{InfluxDBConfig: config}
}

func (c *Controller) CreateControllers(controllers []*DaoClusterTypes.Controller) error {
	controllerRepo := RepoInfluxCluster.NewControllerRepository(&c.InfluxDBConfig)
	err := controllerRepo.CreateControllers(controllers)
	if err != nil {
		scope.Error(err.Error())
		return err
	}
	return nil
}

func (c *Controller) ListControllers(request DaoClusterTypes.ListControllersRequest) ([]*DaoClusterTypes.Controller, error) {
	controllerRepo := RepoInfluxCluster.NewControllerRepository(&c.InfluxDBConfig)
	controllers, err := controllerRepo.ListControllers(request)
	if err != nil {
		scope.Error(err.Error())
		return make([]*DaoClusterTypes.Controller, 0), err
	}
	return controllers, nil
}

func (c *Controller) DeleteControllers(in *ApiResources.DeleteControllersRequest) error {
	controllerRepository := RepoInfluxCluster.NewControllerRepository(&c.InfluxDBConfig)
	return controllerRepository.DeleteControllers(in)
}
