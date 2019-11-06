package responses

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type ApplicationExtended struct {
	*DaoClusterTypes.Application
}

func (n *ApplicationExtended) ProduceApplication() *ApiResources.Application {
	application := &ApiResources.Application{
		ObjectMeta: NewObjectMeta(n.ObjectMeta),
	}
	return application
}

type NamespaceExtended struct {
	*DaoClusterTypes.Namespace
}

func (n *NamespaceExtended) ProduceNamespace() *ApiResources.Namespace {
	namespace := &ApiResources.Namespace{
		ObjectMeta: NewObjectMeta(n.ObjectMeta),
	}
	return namespace
}
