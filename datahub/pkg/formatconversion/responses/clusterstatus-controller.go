package responses

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type ControllerExtended struct {
	*types.Controller
}

func (p *ControllerExtended) ProduceController() *resources.Controller {
	controller := resources.Controller{}
	controller.ObjectMeta = NewObjectMeta(p.ObjectMeta)
	controller.Kind = resources.Kind(resources.Kind_value[p.Kind])
	controller.Replicas = p.Replicas
	controller.SpecReplicas = p.SpecReplicas
	controller.AlamedaControllerSpec = NewAlamedaControllerSpec(p.AlamedaControllerSpec)
	return &controller
}

func NewController(controller *types.Controller) *resources.Controller {
	ctl := resources.Controller{}
	ctl.ObjectMeta = NewObjectMeta(controller.ObjectMeta)
	ctl.Kind = resources.Kind(resources.Kind_value[controller.Kind])
	ctl.Replicas = controller.Replicas
	ctl.SpecReplicas = controller.SpecReplicas
	ctl.AlamedaControllerSpec = NewAlamedaControllerSpec(controller.AlamedaControllerSpec)
	return &ctl
}
