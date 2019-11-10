package responses

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type ApplicationExtended struct {
	*types.Application
}

func (n *ApplicationExtended) ProduceApplication() *resources.Application {
	application := &resources.Application{
		ObjectMeta: NewObjectMeta(n.ObjectMeta),
	}
	return application
}
