package influxdb

import (
	RepoInfluxPlanning "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/plannings"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
)

type AppPlannings struct {
	InfluxDBConfig InternalInflux.Config
}

func NewAppPlanningsWithConfig(config InternalInflux.Config) *AppPlannings {
	return &AppPlannings{InfluxDBConfig: config}
}

func (c *AppPlannings) CreatePlannings(in *ApiPlannings.CreateApplicationPlanningsRequest) error {
	repository := RepoInfluxPlanning.NewAppRepository(&c.InfluxDBConfig)
	return repository.CreatePlannings(in)
}

func (c *AppPlannings) ListPlannings(in *ApiPlannings.ListApplicationPlanningsRequest) ([]*ApiPlannings.ApplicationPlanning, error) {
	repository := RepoInfluxPlanning.NewAppRepository(&c.InfluxDBConfig)
	return repository.ListPlannings(in)
}
