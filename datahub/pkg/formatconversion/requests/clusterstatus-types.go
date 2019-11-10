package requests

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

func NewAlamedaControllerSpec(controllerSpec *resources.AlamedaControllerSpec) types.AlamedaControllerSpec {
	spec := types.AlamedaControllerSpec{
		AlamedaScaler:   NewObjectMeta(controllerSpec.GetAlamedaScaler()),
		Policy:          controllerSpec.GetPolicy().String(),
		EnableExecution: controllerSpec.GetEnableRecommendationExecution(),
	}
	return spec
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
