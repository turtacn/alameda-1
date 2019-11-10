package responses

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

func NewAlamedaControllerSpec(controlSpec types.AlamedaControllerSpec) *resources.AlamedaControllerSpec {
	ctlSpec := resources.AlamedaControllerSpec{}
	ctlSpec.AlamedaScaler = NewObjectMeta(controlSpec.AlamedaScaler)
	ctlSpec.Policy = resources.RecommendationPolicy(resources.RecommendationPolicy_value[controlSpec.Policy])
	ctlSpec.EnableRecommendationExecution = controlSpec.EnableExecution
	return &ctlSpec
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
