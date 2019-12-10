package requests

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	Metadata "github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type CreatePodsRequestExtended struct {
	ApiResources.CreatePodsRequest
}

type ListPodsRequestExtended struct {
	*ApiResources.ListPodsRequest
}

type DeletePodsRequestExtended struct {
	*ApiResources.DeletePodsRequest
}

func (p *CreatePodsRequestExtended) Validate() error {
	return nil
}

func NewPod(pod *ApiResources.Pod) *DaoClusterTypes.Pod {
	if pod != nil {
		// Normalize request
		objectMeta := NewObjectMeta(pod.GetObjectMeta())

		p := DaoClusterTypes.Pod{}
		p.Containers = make([]*DaoClusterTypes.Container, 0)
		p.ObjectMeta = &objectMeta
		p.ResourceLink = pod.GetResourceLink()
		p.CreateTime = pod.GetStartTime()
		p.TopController = NewController(pod.GetTopController())
		p.Status = NewPodStatus(pod.GetStatus())
		p.AppName = pod.GetAppName()
		p.AppPartOf = pod.GetAppPartOf()
		p.AlamedaPodSpec = NewAlamedaPodSpec(pod.GetAlamedaPodSpec())
		for _, container := range pod.GetContainers() {
			p.Containers = append(p.Containers, NewContainer(pod, container))
		}

		return &p
	}
	return nil
}

func NewContainer(pod *ApiResources.Pod, container *ApiResources.Container) *DaoClusterTypes.Container {
	cnt := &DaoClusterTypes.Container{}
	cnt.Name = container.GetName()
	if pod.GetObjectMeta() != nil {
		cnt.PodName = pod.GetObjectMeta().GetName()
		cnt.Namespace = pod.GetObjectMeta().GetNamespace()
		cnt.NodeName = pod.GetObjectMeta().GetNodeName()
		cnt.ClusterName = pod.GetObjectMeta().GetClusterName()
	}
	if pod.GetTopController() != nil {
		cnt.TopControllerName = pod.GetTopController().GetObjectMeta().GetName()
		cnt.TopControllerKind = pod.GetTopController().GetKind().String()
	}
	if pod.GetAlamedaPodSpec() != nil {
		cnt.AlamedaScalerName = pod.GetAlamedaPodSpec().GetAlamedaScaler().GetName()
		cnt.AlamedaScalerScalingTool = pod.GetAlamedaPodSpec().GetScalingTool().String()
	}
	cnt.Resources = NewResourceRequirements(container.GetResources())
	cnt.Status = NewContainerStatus(container.GetStatus())
	return cnt
}

func (p *CreatePodsRequestExtended) ProducePods() []*DaoClusterTypes.Pod {
	pods := make([]*DaoClusterTypes.Pod, 0)
	for _, p := range p.GetPods() {
		pods = append(pods, NewPod(p))
	}
	return pods
}

func (p *ListPodsRequestExtended) Validate() error {
	return nil
}

func (p *ListPodsRequestExtended) ProduceRequest() *DaoClusterTypes.ListPodsRequest {
	request := DaoClusterTypes.NewListPodsRequest()
	request.QueryCondition = QueryConditionExtend{p.GetQueryCondition()}.QueryCondition()
	request.Kind = p.GetKind().String()
	request.ScalingTool = p.GetScalingTool().String()
	if p.GetObjectMeta() != nil {
		for _, meta := range p.GetObjectMeta() {
			// Normalize request
			objectMeta := NewObjectMeta(meta)

			if objectMeta.IsEmpty() {
				request.ObjectMeta = make([]*Metadata.ObjectMeta, 0)
				return request
			}
			request.ObjectMeta = append(request.ObjectMeta, &objectMeta)
		}
	}
	return request
}

func (p *DeletePodsRequestExtended) Validate() error {
	return nil
}

func (p *DeletePodsRequestExtended) ProduceRequest() *DaoClusterTypes.DeletePodsRequest {
	request := DaoClusterTypes.NewDeletePodsRequest()
	if p.GetObjectMeta() != nil {
		for _, meta := range p.GetObjectMeta() {
			// Normalize request
			objectMeta := NewObjectMeta(meta)

			if objectMeta.IsEmpty() {
				request.PodObjectMeta = make([]*DaoClusterTypes.PodObjectMeta, 0)
				return request
			}
			request.PodObjectMeta = append(request.PodObjectMeta, DaoClusterTypes.NewPodObjectMeta(&objectMeta, nil, nil, "", ""))
		}
	}
	return request
}
