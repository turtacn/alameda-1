package influxdb

import (
	RepoInfluxPlanning "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/plannings"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
)

type ControllerPlannings struct {
	InfluxDBConfig InternalInflux.Config
}

func NewControllerPlanningsWithConfig(config InternalInflux.Config) *ControllerPlannings {
	return &ControllerPlannings{InfluxDBConfig: config}
}

func (c *ControllerPlannings) AddControllerPlannings(in *ApiPlannings.CreateControllerPlanningsRequest) error {
	controllerRepository := RepoInfluxPlanning.NewControllerRepository(&c.InfluxDBConfig)
	return controllerRepository.CreateControllerPlannings(in)
}

func (c *ControllerPlannings) ListControllerPlannings(in *ApiPlannings.ListControllerPlanningsRequest) ([]*ApiPlannings.ControllerPlanning, error) {
	controllerRepository := RepoInfluxPlanning.NewControllerRepository(&c.InfluxDBConfig)
	return controllerRepository.ListControllerPlannings(in)
}
