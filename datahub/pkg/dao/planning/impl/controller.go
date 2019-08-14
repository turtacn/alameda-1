package impl

import (
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	RepoInfluxPlanning "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb/planning"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

type Controller struct {
	InfluxDBConfig RepoInflux.Config
}

func (c *Controller) AddControllerPlannings(controllerPlannings []*DatahubV1alpha1.ControllerPlanning) error {
	controllerRepository := RepoInfluxPlanning.NewControllerRepository(&c.InfluxDBConfig)
	return controllerRepository.CreateControllerPlannings(controllerPlannings)
}

func (c *Controller) ListControllerPlannings(in *DatahubV1alpha1.ListControllerPlanningsRequest) ([]*DatahubV1alpha1.ControllerPlanning, error) {
	controllerRepository := RepoInfluxPlanning.NewControllerRepository(&c.InfluxDBConfig)
	return controllerRepository.ListControllerPlannings(in)
}
