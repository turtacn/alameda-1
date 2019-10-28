package influxdb

import (
	RepoInfluxClusterStatus "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/clusterstatus"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type Controller struct {
	InfluxDBConfig InternalInflux.Config
}

func (c *Controller) CreateControllers(controllers []*ApiResources.Controller) error {
	controllerRepository := RepoInfluxClusterStatus.NewControllerRepository(&c.InfluxDBConfig)
	return controllerRepository.CreateControllers(controllers)
}

func (c *Controller) ListControllers(in *ApiResources.ListControllersRequest) ([]*ApiResources.Controller, error) {
	controllerRepository := RepoInfluxClusterStatus.NewControllerRepository(&c.InfluxDBConfig)
	return controllerRepository.ListControllers(in)
}

func (c *Controller) DeleteControllers(in *ApiResources.DeleteControllersRequest) error {
	controllerRepository := RepoInfluxClusterStatus.NewControllerRepository(&c.InfluxDBConfig)
	return controllerRepository.DeleteControllers(in)
}
