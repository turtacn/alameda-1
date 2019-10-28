package influxdb

import (
	RepoInfluxClusterStatus "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/clusterstatus"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

var (
	scope = Log.RegisterScope("dao_implement", "dao implement", 0)
)

// Implement ContainerOperation interface
type Container struct {
	InfluxDBConfig InternalInflux.Config
}

func (container *Container) AddPods(pods []*ApiResources.Pod) error {
	containerRepository := RepoInfluxClusterStatus.NewContainerRepository(&container.InfluxDBConfig)
	return containerRepository.CreateContainers(pods)
}

func (container *Container) DeletePods(pods []*ApiResources.Pod) error {
	containerRepository := RepoInfluxClusterStatus.NewContainerRepository(&container.InfluxDBConfig)
	return containerRepository.DeleteContainers(pods)
}

func (container *Container) ListAlamedaPods(ns, name string, kind ApiResources.Kind, timeRange *ApiCommon.TimeRange) ([]*ApiResources.Pod, error) {
	containerRepository := RepoInfluxClusterStatus.NewContainerRepository(&container.InfluxDBConfig)
	return containerRepository.ListAlamedaContainers(ns, name, kind, timeRange)
}
