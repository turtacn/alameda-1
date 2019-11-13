package responses

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type PodExtended struct {
	*types.Pod
}

func (n *PodExtended) ProducePod() *resources.Pod {
	pod := resources.Pod{}
	pod.ObjectMeta = NewObjectMeta(*n.ObjectMeta)
	pod.StartTime = n.CreateTime
	pod.ResourceLink = n.ResourceLink
	pod.AppName = n.AppName
	pod.AppPartOf = n.AppPartOf
	pod.Containers = make([]*resources.Container, 0)
	for _, cnt := range n.Containers {
		container := resources.Container{}
		container.Name = cnt.Name
		container.Resources = NewResourceRequirements(cnt.Resources)
		container.Status = NewContainerStatus(cnt.Status)
		pod.Containers = append(pod.Containers, &container)
	}
	pod.TopController = NewController(n.TopController)
	pod.Status = NewPodStatus(n.Status)
	pod.AlamedaPodSpec = NewAlamedaPodSpec(n.AlamedaPodSpec)
	return &pod
}
