package influxdb

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/clusterstatus"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

// Implement ContainerOperation interface
type Pod struct {
	InfluxDBConfig InternalInflux.Config
}

func NewPodWithConfig(config InternalInflux.Config) DaoClusterTypes.PodDAO {
	return &Pod{InfluxDBConfig: config}
}

func (p *Pod) CreatePods(pods []*DaoClusterTypes.Pod) error {
	delContainerReq := DaoClusterTypes.NewDeleteContainersRequest()
	for _, pod := range pods {
		containerMeta := DaoClusterTypes.ContainerObjectMeta{}
		containerMeta.PodName = pod.ObjectMeta.Name
		containerMeta.Namespace = pod.ObjectMeta.Namespace
		containerMeta.NodeName = pod.ObjectMeta.NodeName
		containerMeta.ClusterName = pod.ObjectMeta.ClusterName
		containerMeta.TopControllerName = pod.TopController.ObjectMeta.Name
		containerMeta.TopControllerKind = pod.TopController.Kind
		containerMeta.AlamedaScalerName = pod.AlamedaPodSpec.AlamedaScaler.Name
		containerMeta.AlamedaScalerScalingTool = pod.AlamedaPodSpec.ScalingTool
		delContainerReq.ContainerObjectMeta = append(delContainerReq.ContainerObjectMeta, &containerMeta)
	}

	containerMap := make(map[string][]*DaoClusterTypes.Container)
	for _, pod := range pods {
		identifier := pod.ClusterNamespacePodName()
		containerMap[identifier] = make([]*DaoClusterTypes.Container, 0)
		for _, container := range pod.Containers {
			containerMap[identifier] = append(containerMap[identifier], container)
		}
	}

	// Do delete containers before creating them
	containerRepo := RepoInfluxCluster.NewContainerRepository(p.InfluxDBConfig)
	err := containerRepo.DeleteContainers(delContainerReq)
	if err != nil {
		scope.Error("failed to delete container in influxdb when creating pods")
		return err
	}

	// Create containers
	if err := containerRepo.CreateContainers(containerMap); err != nil {
		scope.Error(err.Error())
		return err
	}

	// Create pods
	podRepo := RepoInfluxCluster.NewPodRepository(p.InfluxDBConfig)
	if err := podRepo.CreatePods(pods); err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}

func (p *Pod) ListPods(request *DaoClusterTypes.ListPodsRequest) ([]*DaoClusterTypes.Pod, error) {
	podRepo := RepoInfluxCluster.NewPodRepository(p.InfluxDBConfig)
	pods, err := podRepo.ListPods(request)
	if err != nil {
		scope.Error(err.Error())
		return make([]*DaoClusterTypes.Pod, 0), err
	}

	containerRequest := DaoClusterTypes.NewListContainersRequest()
	for _, pod := range pods {
		containerMeta := DaoClusterTypes.ContainerObjectMeta{}
		containerMeta.PodName = pod.ObjectMeta.Name
		containerMeta.Namespace = pod.ObjectMeta.Namespace
		containerMeta.NodeName = pod.ObjectMeta.NodeName
		containerMeta.ClusterName = pod.ObjectMeta.ClusterName
		containerMeta.TopControllerName = pod.TopController.ObjectMeta.Name
		containerMeta.TopControllerKind = pod.TopController.Kind
		containerMeta.AlamedaScalerName = pod.AlamedaPodSpec.AlamedaScaler.Name
		containerMeta.AlamedaScalerScalingTool = pod.AlamedaPodSpec.ScalingTool
		containerRequest.ContainerObjectMeta = append(containerRequest.ContainerObjectMeta, &containerMeta)
	}

	containerRepo := RepoInfluxCluster.NewContainerRepository(p.InfluxDBConfig)
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

func (p *Pod) DeletePods(request *DaoClusterTypes.DeletePodsRequest) error {
	delContainerReq := DaoClusterTypes.NewDeleteContainersRequest()
	for _, podObjectMeta := range request.PodObjectMeta {
		containerMeta := DaoClusterTypes.ContainerObjectMeta{}
		containerMeta.TopControllerKind = podObjectMeta.Kind
		containerMeta.AlamedaScalerScalingTool = podObjectMeta.ScalingTool
		if podObjectMeta.ObjectMeta != nil {
			containerMeta.PodName = podObjectMeta.ObjectMeta.Name
			containerMeta.Namespace = podObjectMeta.ObjectMeta.Namespace
			containerMeta.NodeName = podObjectMeta.ObjectMeta.NodeName
			containerMeta.ClusterName = podObjectMeta.ObjectMeta.ClusterName
		}
		if podObjectMeta.TopController != nil {
			containerMeta.TopControllerName = podObjectMeta.TopController.Name
			if containerMeta.Namespace == "" {
				containerMeta.Namespace = podObjectMeta.TopController.Namespace
			}
			if containerMeta.ClusterName == "" {
				containerMeta.ClusterName = podObjectMeta.TopController.ClusterName
			}
		}
		if podObjectMeta.AlamedaScaler != nil {
			containerMeta.AlamedaScalerName = podObjectMeta.AlamedaScaler.Name
			if containerMeta.Namespace == "" {
				containerMeta.Namespace = podObjectMeta.AlamedaScaler.Namespace
			}
			if containerMeta.ClusterName == "" {
				containerMeta.ClusterName = podObjectMeta.AlamedaScaler.ClusterName
			}
		}
		delContainerReq.ContainerObjectMeta = append(delContainerReq.ContainerObjectMeta, &containerMeta)
	}

	// Delete pods
	podRepo := RepoInfluxCluster.NewPodRepository(p.InfluxDBConfig)
	if err := podRepo.DeletePods(request); err != nil {
		scope.Error(err.Error())
		return err
	}

	// Delete containers
	containerRepo := RepoInfluxCluster.NewContainerRepository(p.InfluxDBConfig)
	if err := containerRepo.DeleteContainers(delContainerReq); err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}
