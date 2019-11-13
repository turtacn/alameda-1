package requests

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type CreatePodsRequestExtended struct {
	ApiResources.CreatePodsRequest
}

type ListPodsRequestExtended struct {
	*ApiResources.ListPodsRequest
}

func (r *CreatePodsRequestExtended) Validate() error {
	return nil
}

func (r *CreatePodsRequestExtended) ProducePods() []*DaoClusterTypes.Pod {
	pods := make([]*DaoClusterTypes.Pod, 0)

	for _, p := range r.GetPods() {
		objectMeta := NewObjectMeta(p.GetObjectMeta())

		pod := DaoClusterTypes.NewPod()
		pod.ObjectMeta = &objectMeta
		pod.ResourceLink = p.ResourceLink
		for _, container := range p.GetContainers() {
			cnt := &DaoClusterTypes.Container{}
			cnt.Name = container.Name
			cnt.PodName = pod.ObjectMeta.Name
			cnt.Namespace = pod.ObjectMeta.Namespace
			cnt.NodeName = pod.ObjectMeta.NodeName
			cnt.ClusterName = pod.ObjectMeta.ClusterName
			cnt.Resources = NewResourceRequirements(container.GetResources())
			cnt.Status = NewContainerStatus(container.GetStatus())
			pod.Containers = append(pod.Containers, cnt)
		}
		pod.CreateTime = p.GetStartTime()
		pod.TopController = NewController(p.GetTopController())
		pod.Status = NewPodStatus(p.GetStatus())
		pod.AppName = p.GetAppName()
		pod.AppPartOf = p.GetAppPartOf()
		pod.AlamedaPodSpec = NewAlamedaPodSpec(p.GetAlamedaPodSpec())

		pods = append(pods, pod)
	}

	return pods
}

func (r *ListPodsRequestExtended) Validate() error {
	return nil
}

func (r *ListPodsRequestExtended) ProduceRequest() DaoClusterTypes.ListPodsRequest {
	request := DaoClusterTypes.NewListPodsRequest()
	request.QueryCondition = QueryConditionExtend{r.GetQueryCondition()}.QueryCondition()
	if r.GetObjectMeta() != nil {
		for _, meta := range r.GetObjectMeta() {
			// Normalize request
			objectMeta := NewObjectMeta(meta)

			if objectMeta.IsEmpty() {
				request := DaoClusterTypes.NewListPodsRequest()
				request.Kind = r.GetKind().String()
				request.ScalingTool = r.GetScalingTool().String()
				return request
			}
			request.ObjectMeta = append(request.ObjectMeta, objectMeta)
		}
	}
	request.Kind = r.GetKind().String()
	request.ScalingTool = r.GetScalingTool().String()
	return request
}
