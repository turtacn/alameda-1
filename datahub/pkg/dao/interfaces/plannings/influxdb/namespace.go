package influxdb

import (
	RepoInfluxPlanning "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/plannings"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
)

type NamespacePlannings struct {
	InfluxDBConfig InternalInflux.Config
}

func NewNamespacePlanningsWithConfig(config InternalInflux.Config) *NamespacePlannings {
	return &NamespacePlannings{InfluxDBConfig: config}
}

func (c *NamespacePlannings) CreatePlannings(in *ApiPlannings.CreateNamespacePlanningsRequest) error {
	repository := RepoInfluxPlanning.NewNamespaceRepository(&c.InfluxDBConfig)
	return repository.CreatePlannings(in)
}

func (c *NamespacePlannings) ListPlannings(in *ApiPlannings.ListNamespacePlanningsRequest) ([]*ApiPlannings.NamespacePlanning, error) {
	repository := RepoInfluxPlanning.NewNamespaceRepository(&c.InfluxDBConfig)
	return repository.ListPlannings(in)
}
