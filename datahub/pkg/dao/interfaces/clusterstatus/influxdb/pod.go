package influxdb

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/clusterstatus"
	//Metadata "github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	//ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	//ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

// Implement ContainerOperation interface
type Pod struct {
	InfluxDBConfig InternalInflux.Config
}

func NewPodWithConfig(config InternalInflux.Config) DaoClusterTypes.PodDAO {
	return &Pod{InfluxDBConfig: config}
}

func (p *Pod) CreatePods(pods []*DaoClusterTypes.Pod) error {
	podRepo := RepoInfluxCluster.NewPodRepository(&p.InfluxDBConfig)
	if err := podRepo.CreatePods(pods); err != nil {
		scope.Error(err.Error())
		return err
	}

	containerMap := make(map[string][]*DaoClusterTypes.Container)
	for _, pod := range pods {
		identifier := pod.ClusterNamespacePodName()
		containerMap[identifier] = make([]*DaoClusterTypes.Container, 0)
		for _, container := range pod.Containers {
			containerMap[identifier] = append(containerMap[identifier], container)
		}
	}

	containerRepo := RepoInfluxCluster.NewContainerRepository(&p.InfluxDBConfig)
	if err := containerRepo.CreateContainers(containerMap); err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}

func (p *Pod) ListPods(request DaoClusterTypes.ListPodsRequest) ([]*DaoClusterTypes.Pod, error) {
	podRepo := RepoInfluxCluster.NewPodRepository(&p.InfluxDBConfig)
	pods, err := podRepo.ListPods(request)
	if err != nil {
		scope.Error(err.Error())
		return make([]*DaoClusterTypes.Pod, 0), err
	}

	containerRequest := DaoClusterTypes.NewListContainersRequest()
	for _, pod := range pods {
		containerMeta := DaoClusterTypes.ContainerObjectMeta{}
		containerMeta.PodName = pod.ObjectMeta.Name
		containerMeta.ObjectMeta.Namespace = pod.ObjectMeta.Namespace
		containerMeta.ObjectMeta.NodeName = pod.ObjectMeta.NodeName
		containerMeta.ObjectMeta.ClusterName = pod.ObjectMeta.ClusterName
		containerRequest.ContainerObjectMeta = append(containerRequest.ContainerObjectMeta, containerMeta)
	}

	containerRepo := RepoInfluxCluster.NewContainerRepository(&p.InfluxDBConfig)
	containerMap, err := containerRepo.ListContainers(containerRequest)
	for clusterNamespaceNodeName, containers := range containerMap {
		for _, pod := range pods {
			if pod.ClusterNamespacePodName() == clusterNamespaceNodeName {
				for _, container := range containers {
					pod.Containers = append(pod.Containers, container)
				}
				break
			}
		}
	}

	return pods, nil
}

/*func (p *Pod) DeletePods(pods []*ApiResources.Pod) error {
	containerRepository := RepoInfluxCluster.NewContainerRepository(&container.InfluxDBConfig)
	return containerRepository.DeleteContainers(pods)
}*/
