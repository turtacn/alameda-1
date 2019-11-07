package requests

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type CreateNamespacesRequestExtended struct {
	ApiResources.CreateNamespacesRequest
}

type ListNamespacesRequestExtended struct {
	*ApiResources.ListNamespacesRequest
}

func (r *CreateNamespacesRequestExtended) Validate() error {
	return nil
}

func (r *CreateNamespacesRequestExtended) ProduceNamespaces() []*DaoClusterTypes.Namespace {
	namespaces := make([]*DaoClusterTypes.Namespace, 0)

	for _, ns := range r.GetNamespaces() {
		// Normalize request
		objectMeta := NewObjectMeta(ns.GetObjectMeta())
		objectMeta.Namespace = ""
		objectMeta.NodeName = ""

		namespace := DaoClusterTypes.NewNamespace()
		namespace.ObjectMeta = objectMeta
		namespaces = append(namespaces, namespace)
	}

	return namespaces
}

func (r *ListNamespacesRequestExtended) Validate() error {
	return nil
}

func (r *ListNamespacesRequestExtended) ProduceRequest() DaoClusterTypes.ListNamespacesRequest {
	request := DaoClusterTypes.NewListNamespacesRequest()
	if r.GetObjectMeta() != nil {
		for _, meta := range r.GetObjectMeta() {
			// Normalize request
			objectMeta := NewObjectMeta(meta)
			objectMeta.Namespace = ""
			objectMeta.NodeName = ""

			if objectMeta.IsEmpty() {
				return DaoClusterTypes.NewListNamespacesRequest()
			}
			request.ObjectMeta = append(request.ObjectMeta, objectMeta)
		}
	}
	return request
}
