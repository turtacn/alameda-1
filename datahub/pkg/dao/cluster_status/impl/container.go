package impl

import (
	RepoInfluxClusterStatus "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb/cluster_status"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

var (
	scope = log.RegisterScope("dao_implement", "dao implement", 0)
)

// Implement ContainerOperation interface
type Container struct {
	InfluxDBConfig InternalInflux.Config
}

func (container *Container) AddPods(pods []*datahub_v1alpha1.Pod) error {
	containerRepository := RepoInfluxClusterStatus.NewContainerRepository(&container.InfluxDBConfig)
	return containerRepository.CreateContainers(pods)
}

func (container *Container) DeletePods(pods []*datahub_v1alpha1.Pod) error {
	containerRepository := RepoInfluxClusterStatus.NewContainerRepository(&container.InfluxDBConfig)
	return containerRepository.DeleteContainers(pods)
}

func (container *Container) ListAlamedaPods(ns, name string, kind datahub_v1alpha1.Kind, timeRange *datahub_v1alpha1.TimeRange) ([]*datahub_v1alpha1.Pod, error) {
	containerRepository := RepoInfluxClusterStatus.NewContainerRepository(&container.InfluxDBConfig)
	return containerRepository.ListAlamedaContainers(ns, name, kind, timeRange)
}
