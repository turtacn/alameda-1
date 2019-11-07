package requests

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

func NewObjectMeta(objectMeta *resources.ObjectMeta) metadata.ObjectMeta {
	meta := metadata.ObjectMeta{
		Name:        objectMeta.GetName(),
		Namespace:   objectMeta.GetNamespace(),
		NodeName:    objectMeta.GetNodeName(),
		ClusterName: objectMeta.GetClusterName(),
		Uid:         objectMeta.GetUid(),
	}
	return meta
}

func NewOwnerReference(ownerReference *resources.OwnerReference) types.OwnerReference {
	owner := types.OwnerReference{
		ObjectMeta: NewObjectMeta(ownerReference.GetObjectMeta()),
		Kind:       ownerReference.GetKind().String(),
	}
	return owner
}

func NewAlamedaControllerSpec(controllerSpec *resources.AlamedaControllerSpec) types.AlamedaControllerSpec {
	spec := types.AlamedaControllerSpec{
		AlamedaScaler:   NewObjectMeta(controllerSpec.GetAlamedaScaler()),
		Policy:          controllerSpec.GetPolicy().String(),
		EnableExecution: controllerSpec.GetEnableRecommendationExecution(),
	}
	return spec
}
