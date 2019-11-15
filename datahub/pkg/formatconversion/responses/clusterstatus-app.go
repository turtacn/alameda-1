package responses

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type ApplicationExtended struct {
	*types.Application
}

func (n *ApplicationExtended) ProduceApplication() *resources.Application {
	application := &resources.Application{}
	application.ObjectMeta = NewObjectMeta(n.ObjectMeta)
	application.AlamedaApplicationSpec = NewAlamedaApplicationSpec(n.AlamedaApplicationSpec)
	if n.Controllers != nil {
		application.Controllers = make([]*resources.Controller, 0)
		for _, controller := range n.Controllers {
			application.Controllers = append(application.Controllers, NewController(controller))
		}
	}
	return application
}
