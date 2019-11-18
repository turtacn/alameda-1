package influxdb

import (
	RepoInfluxPlanning "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/plannings"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
)

type ClusterPlannings struct {
	InfluxDBConfig InternalInflux.Config
}

func NewClusterPlanningsWithConfig(config InternalInflux.Config) *ClusterPlannings {
	return &ClusterPlannings{InfluxDBConfig: config}
}

func (c *ClusterPlannings) CreatePlannings(in *ApiPlannings.CreateClusterPlanningsRequest) error {
	repository := RepoInfluxPlanning.NewClusterRepository(&c.InfluxDBConfig)
	return repository.CreatePlannings(in)
}

func (c *ClusterPlannings) ListPlannings(in *ApiPlannings.ListClusterPlanningsRequest) ([]*ApiPlannings.ClusterPlanning, error) {
	repository := RepoInfluxPlanning.NewClusterRepository(&c.InfluxDBConfig)
	return repository.ListPlannings(in)
}
