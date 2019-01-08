package impl

import (
	influxdb_repository "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	influxdb_repository_cluster_status "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb/cluster_status"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

var (
	containerImplScope = log.RegisterScope("container_dao_implement", "container dao implement", 0)
)

// Implement ContainerOperation interface
type Container struct {
	InfluxDBConfig influxdb_repository.Config
}

func (container *Container) AddPods(pods []*datahub_v1alpha1.Pod) error {
	containerRepository := influxdb_repository_cluster_status.NewContainerRepository(&container.InfluxDBConfig)
	return containerRepository.CreateContainers(pods)
}

func (container *Container) DeletePods(pods []*datahub_v1alpha1.Pod) error {
	containerRepository := influxdb_repository_cluster_status.NewContainerRepository(&container.InfluxDBConfig)
	return containerRepository.DeleteContainers(pods)
}

func (container *Container) ListAlamedaPods(scalerNS, scalerName string) ([]*datahub_v1alpha1.Pod, error) {
	containerRepository := influxdb_repository_cluster_status.NewContainerRepository(&container.InfluxDBConfig)
	return containerRepository.ListAlamedaContainers(scalerNS, scalerName)
}
