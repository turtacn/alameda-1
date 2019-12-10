package responses

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type NamespaceExtended struct {
	*types.Namespace
}

func (p *NamespaceExtended) ProduceNamespace() *resources.Namespace {
	namespace := &resources.Namespace{
		ObjectMeta: NewObjectMeta(p.ObjectMeta),
	}
	return namespace
}
