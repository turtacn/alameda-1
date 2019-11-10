package responses

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type ControllerExtended struct {
	*types.Controller
}

func (n *ControllerExtended) ProduceController() *resources.Controller {
	controller := resources.Controller{}
	controller.OwnerReferences = make([]*resources.OwnerReference, 0)
	controller.ObjectMeta = NewObjectMeta(n.ObjectMeta)
	controller.Kind = resources.Kind(resources.Kind_value[n.Kind])
	controller.Replicas = n.Replicas
	controller.SpecReplicas = n.SpecReplicas
	controller.AlamedaControllerSpec = NewAlamedaControllerSpec(n.AlamedaControllerSpec)
	for _, ownerReference := range n.OwnerReferences {
		controller.OwnerReferences = append(controller.OwnerReferences, NewOwnerReference(ownerReference))
	}
	return &controller
}
