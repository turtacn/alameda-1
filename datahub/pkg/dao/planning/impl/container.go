package impl

import (
	RepoInflux "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	RepoInfluxPlanning "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb/planning"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

var (
	scope = Log.RegisterScope("planning_dao_implement", "planning dao implement", 0)
)

// Container Implements ContainerOperation interface
type Container struct {
	InfluxDBConfig RepoInflux.Config
}

// AddPodPlannings add pod plannings to database
func (container *Container) AddPodPlannings(in *DatahubV1alpha1.CreatePodPlanningsRequest) error {
	containerRepository := RepoInfluxPlanning.NewContainerRepository(&container.InfluxDBConfig)
	return containerRepository.CreateContainerPlannings(in)
}

// ListPodPlannings list pod plannings
func (container *Container) ListPodPlannings(in *DatahubV1alpha1.ListPodPlanningsRequest) ([]*DatahubV1alpha1.PodPlanning, error) {
	containerRepository := RepoInfluxPlanning.NewContainerRepository(&container.InfluxDBConfig)
	return containerRepository.ListContainerPlannings(in)
}
