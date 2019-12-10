package responses

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

func NewAlamedaPodSpec(podSpec *types.AlamedaPodSpec) *resources.AlamedaPodSpec {
	if podSpec != nil {
		spec := resources.AlamedaPodSpec{}
		if podSpec.AlamedaScaler != nil {
			spec.AlamedaScaler = NewObjectMeta(podSpec.AlamedaScaler)
		}
		spec.UsedRecommendationId = podSpec.UsedRecommendationId
		spec.ScalingTool = resources.ScalingTool(resources.ScalingTool_value[podSpec.ScalingTool])
		spec.Policy = resources.RecommendationPolicy(resources.RecommendationPolicy_value[podSpec.Policy])
		if podSpec.AlamedaScalerResources != nil {
			spec.AlamedaScalerResources = NewResourceRequirements(podSpec.AlamedaScalerResources)
		}
		return &spec
	}
	return nil
}

func NewAlamedaControllerSpec(controlSpec *types.AlamedaControllerSpec) *resources.AlamedaControllerSpec {
	if controlSpec != nil {
		ctlSpec := resources.AlamedaControllerSpec{}
		ctlSpec.AlamedaScaler = NewObjectMeta(controlSpec.AlamedaScaler)
		ctlSpec.ScalingTool = resources.ScalingTool(resources.ScalingTool_value[controlSpec.ScalingTool])
		ctlSpec.Policy = resources.RecommendationPolicy(resources.RecommendationPolicy_value[controlSpec.Policy])
		ctlSpec.EnableRecommendationExecution = controlSpec.EnableExecution
		return &ctlSpec
	}
	return nil
}

func NewAlamedaApplicationSpec(applicationSpec *types.AlamedaApplicationSpec) *resources.AlamedaApplicationSpec {
	if applicationSpec != nil {
		spec := resources.AlamedaApplicationSpec{}
		spec.ScalingTool = resources.ScalingTool(resources.ScalingTool_value[applicationSpec.ScalingTool])
		return &spec
	}
	return nil
}

func NewAlamedaNodeSpec(nodeSpec *types.AlamedaNodeSpec) *resources.AlamedaNodeSpec {
	if nodeSpec != nil {
		spec := resources.AlamedaNodeSpec{}
		if nodeSpec.Provider != nil {
			spec.Provider = &resources.Provider{}
			spec.Provider.Provider = nodeSpec.Provider.Provider
			spec.Provider.InstanceType = nodeSpec.Provider.InstanceType
			spec.Provider.Region = nodeSpec.Provider.Region
			spec.Provider.Zone = nodeSpec.Provider.Zone
			spec.Provider.Os = nodeSpec.Provider.Os
			spec.Provider.Role = nodeSpec.Provider.Role
			spec.Provider.InstanceId = nodeSpec.Provider.InstanceId
			spec.Provider.StorageSize = nodeSpec.Provider.StorageSize
		}
		return &spec
	}
	return nil
}

func NewCapacity(capacity *types.Capacity) *resources.Capacity {
	if capacity != nil {
		cp := resources.Capacity{}
		cp.CpuCores = capacity.CpuCores
		cp.MemoryBytes = capacity.MemoryBytes
		cp.NetworkMegabitsPerSecond = capacity.NetworkMegabitsPerSecond
		return &cp
	}
	return nil
}

func NewResourceRequirements(requirements *types.ResourceRequirements) *resources.ResourceRequirements {
	if requirements != nil {
		resourceReq := resources.ResourceRequirements{}

		if requirements.Limits != nil {
			resourceReq.Limits = make(map[int32]string)
			for k, v := range requirements.Limits {
				resourceReq.Limits[k] = v
			}
		}

		if requirements.Requests != nil {
			resourceReq.Requests = make(map[int32]string)
			for k, v := range requirements.Requests {
				resourceReq.Requests[k] = v
			}
		}

		return &resourceReq
	}
	return nil
}
