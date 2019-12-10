package responses

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type PodExtended struct {
	*types.Pod
}

func (p *PodExtended) ProducePod() *resources.Pod {
	pod := resources.Pod{}
	pod.ObjectMeta = NewObjectMeta(p.ObjectMeta)
	pod.StartTime = p.CreateTime
	pod.ResourceLink = p.ResourceLink
	pod.AppName = p.AppName
	pod.AppPartOf = p.AppPartOf
	pod.Containers = make([]*resources.Container, 0)
	for _, cnt := range p.Containers {
		container := resources.Container{}
		container.Name = cnt.Name
		container.Resources = NewResourceRequirements(cnt.Resources)
		container.Status = NewContainerStatus(cnt.Status)
		pod.Containers = append(pod.Containers, &container)
	}
	pod.TopController = NewController(p.TopController)
	pod.Status = NewPodStatus(p.Status)
	pod.AlamedaPodSpec = NewAlamedaPodSpec(p.AlamedaPodSpec)
	return &pod
}
