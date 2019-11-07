package responses

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type ControllerExtended struct {
	*DaoClusterTypes.Controller
}

func (n *ControllerExtended) ProduceController() *ApiResources.Controller {
	controller := ApiResources.Controller{}
	controller.OwnerReferences = make([]*ApiResources.OwnerReference, 0)
	controller.ObjectMeta = NewObjectMeta(n.ObjectMeta)
	controller.Kind = ApiResources.Kind(ApiResources.Kind_value[n.Kind])
	controller.Replicas = n.Replicas
	controller.SpecReplicas = n.SpecReplicas
	controller.AlamedaControllerSpec = NewAlamedaControllerSpec(n.AlamedaControllerSpec)
	for _, ownerReference := range n.OwnerReferences {
		controller.OwnerReferences = append(controller.OwnerReferences, NewOwnerReference(ownerReference))
	}
	return &controller
}
