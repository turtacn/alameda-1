package responses

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type ApplicationExtended struct {
	*types.Application
}

func (p *ApplicationExtended) ProduceApplication() *resources.Application {
	application := &resources.Application{}
	application.ObjectMeta = NewObjectMeta(p.ObjectMeta)
	application.AlamedaApplicationSpec = NewAlamedaApplicationSpec(p.AlamedaApplicationSpec)
	if p.Controllers != nil {
		application.Controllers = make([]*resources.Controller, 0)
		for _, controller := range p.Controllers {
			application.Controllers = append(application.Controllers, NewController(controller))
		}
	}
	return application
}
