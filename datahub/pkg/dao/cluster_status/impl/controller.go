package impl

import (
	RepoInfluxClusterStatus "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb/cluster_status"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

type Controller struct {
	InfluxDBConfig InternalInflux.Config
}

func (c *Controller) CreateControllers(controllers []*datahub_v1alpha1.Controller) error {
	controllerRepository := RepoInfluxClusterStatus.NewControllerRepository(&c.InfluxDBConfig)
	return controllerRepository.CreateControllers(controllers)
}

func (c *Controller) ListControllers(in *datahub_v1alpha1.ListControllersRequest) ([]*datahub_v1alpha1.Controller, error) {
	controllerRepository := RepoInfluxClusterStatus.NewControllerRepository(&c.InfluxDBConfig)
	return controllerRepository.ListControllers(in)
}

func (c *Controller) DeleteControllers(in *datahub_v1alpha1.DeleteControllersRequest) error {
	controllerRepository := RepoInfluxClusterStatus.NewControllerRepository(&c.InfluxDBConfig)
	return controllerRepository.DeleteControllers(in)
}
