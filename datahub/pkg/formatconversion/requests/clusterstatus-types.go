package requests

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

func NewAlamedaControllerSpec(controllerSpec *resources.AlamedaControllerSpec) *types.AlamedaControllerSpec {
	objectMeta := NewObjectMeta(controllerSpec.GetAlamedaScaler())

	spec := types.AlamedaControllerSpec{
		AlamedaScaler:   &objectMeta,
		ScalingTool:     controllerSpec.GetScalingTool().String(),
		Policy:          controllerSpec.GetPolicy().String(),
		EnableExecution: controllerSpec.GetEnableRecommendationExecution(),
	}
	return &spec
}

func NewAlamedaPodSpec(podSpec *resources.AlamedaPodSpec) *types.AlamedaPodSpec {
	if podSpec != nil {
		objectMeta := NewObjectMeta(podSpec.GetAlamedaScaler())

		spec := &types.AlamedaPodSpec{}
		spec.AlamedaScaler = &objectMeta
		spec.Policy = podSpec.GetPolicy().String()
		spec.UsedRecommendationId = podSpec.GetUsedRecommendationId()
		spec.AlamedaScalerResources = NewResourceRequirements(podSpec.GetAlamedaScalerResources())
		spec.ScalingTool = podSpec.GetScalingTool().String()
		return spec
	}
	return nil
}

func NewAlamedaApplicationSpec(applicationSpec *resources.AlamedaApplicationSpec) *types.AlamedaApplicationSpec {
	if applicationSpec != nil {
		spec := &types.AlamedaApplicationSpec{}
		spec.ScalingTool = applicationSpec.GetScalingTool().String()
		return spec
	}
	return nil
}

func NewAlamedaNodeSpec(nodeSpec *resources.AlamedaNodeSpec) *types.AlamedaNodeSpec {
	if nodeSpec != nil {
		spec := &types.AlamedaNodeSpec{}
		if provider := nodeSpec.Provider; provider != nil {
			spec.Provider = &types.Provider{
				Provider:     provider.Provider,
				InstanceType: provider.InstanceType,
				Region:       provider.Region,
				Zone:         provider.Zone,
				Os:           provider.Os,
				Role:         provider.Role,
				InstanceId:   provider.InstanceId,
				StorageSize:  provider.StorageSize,
			}
		}
		return spec
	}
	return nil
}

func NewCapacity(capacity *resources.Capacity) *types.Capacity {
	if capacity != nil {
		c := &types.Capacity{
			CpuCores:                 capacity.CpuCores,
			MemoryBytes:              capacity.MemoryBytes,
			NetworkMegabitsPerSecond: capacity.NetworkMegabitsPerSecond,
		}
		return c
	}
	return nil
}

func NewResourceRequirements(resourceReq *resources.ResourceRequirements) *types.ResourceRequirements {
	if resourceReq != nil {
		requirements := types.ResourceRequirements{}
		if resourceReq.GetLimits() != nil {
			requirements.Limits = make(map[int32]string)
			for k, v := range resourceReq.GetLimits() {
				requirements.Limits[k] = v
			}
		}
		if resourceReq.GetRequests() != nil {
			requirements.Requests = make(map[int32]string)
			for k, v := range resourceReq.GetRequests() {
				requirements.Requests[k] = v
			}
		}
		return &requirements
	}
	return nil
}
