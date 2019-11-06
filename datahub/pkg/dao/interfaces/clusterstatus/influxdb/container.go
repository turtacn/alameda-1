package influxdb

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/clusterstatus"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

// Implement ContainerOperation interface
type Container struct {
	InfluxDBConfig InternalInflux.Config
}

func NewContainerWithConfig(config InternalInflux.Config) DaoClusterTypes.ContainerDAO {
	return &Container{InfluxDBConfig: config}
}

func (container *Container) AddPods(pods []*ApiResources.Pod) error {
	containerRepository := RepoInfluxCluster.NewContainerRepository(&container.InfluxDBConfig)
	return containerRepository.CreateContainers(pods)
}

func (container *Container) DeletePods(pods []*ApiResources.Pod) error {
	containerRepository := RepoInfluxCluster.NewContainerRepository(&container.InfluxDBConfig)
	return containerRepository.DeleteContainers(pods)
}

func (container *Container) ListAlamedaPods(ns, name string, kind ApiResources.Kind, timeRange *ApiCommon.TimeRange) ([]*ApiResources.Pod, error) {
	containerRepository := RepoInfluxCluster.NewContainerRepository(&container.InfluxDBConfig)
	return containerRepository.ListAlamedaContainers(ns, name, kind, timeRange)
}
