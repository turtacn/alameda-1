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

func NewController(controller *types.Controller) *resources.Controller {
	ctl := resources.Controller{}
	ctl.ObjectMeta = NewObjectMeta(controller.ObjectMeta)
	ctl.Kind = resources.Kind(resources.Kind_value[controller.Kind])
	ctl.Replicas = controller.Replicas
	ctl.SpecReplicas = controller.SpecReplicas
	if controller.OwnerReferences != nil && len(controller.OwnerReferences) > 0 {
		ctl.OwnerReferences = make([]*resources.OwnerReference, 0)
		for _, ownerReference := range controller.OwnerReferences {
			ctl.OwnerReferences = append(ctl.OwnerReferences, NewOwnerReference(ownerReference))
		}
	}
	ctl.AlamedaControllerSpec = NewAlamedaControllerSpec(controller.AlamedaControllerSpec)
	return &ctl
}
