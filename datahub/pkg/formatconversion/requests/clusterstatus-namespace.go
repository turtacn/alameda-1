package requests

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	Metadata "github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type CreateNamespacesRequestExtended struct {
	ApiResources.CreateNamespacesRequest
}

type ListNamespacesRequestExtended struct {
	*ApiResources.ListNamespacesRequest
}

type DeleteNamespacesRequestExtended struct {
	*ApiResources.DeleteNamespacesRequest
}

func NewNamespace(namespace *ApiResources.Namespace) *DaoClusterTypes.Namespace {
	if namespace != nil {
		// Normalize request
		objectMeta := NewObjectMeta(namespace.GetObjectMeta())
		objectMeta.Namespace = ""
		objectMeta.NodeName = ""

		ns := DaoClusterTypes.Namespace{}
		ns.ObjectMeta = &objectMeta

		return &ns
	}
	return nil
}

func (p *CreateNamespacesRequestExtended) Validate() error {
	return nil
}

func (p *CreateNamespacesRequestExtended) ProduceNamespaces() []*DaoClusterTypes.Namespace {
	namespaces := make([]*DaoClusterTypes.Namespace, 0)

	for _, ns := range p.GetNamespaces() {
		namespaces = append(namespaces, NewNamespace(ns))
	}

	return namespaces
}

func (p *ListNamespacesRequestExtended) Validate() error {
	return nil
}

func (p *ListNamespacesRequestExtended) ProduceRequest() *DaoClusterTypes.ListNamespacesRequest {
	request := DaoClusterTypes.NewListNamespacesRequest()
	if p.GetObjectMeta() != nil {
		for _, meta := range p.GetObjectMeta() {
			// Normalize request
			objectMeta := NewObjectMeta(meta)
			objectMeta.Namespace = ""
			objectMeta.NodeName = ""

			if objectMeta.IsEmpty() {
				request.ObjectMeta = make([]*Metadata.ObjectMeta, 0)
				return request
			}
			request.ObjectMeta = append(request.ObjectMeta, &objectMeta)
		}
	}
	return request
}

func (p *DeleteNamespacesRequestExtended) Validate() error {
	return nil
}

func (p *DeleteNamespacesRequestExtended) ProduceRequest() *DaoClusterTypes.DeleteNamespacesRequest {
	request := DaoClusterTypes.NewDeleteNamespacesRequest()
	if p.GetObjectMeta() != nil {
		for _, meta := range p.GetObjectMeta() {
			// Normalize request
			objectMeta := NewObjectMeta(meta)
			objectMeta.Namespace = ""
			objectMeta.NodeName = ""

			if objectMeta.IsEmpty() {
				request.ObjectMeta = make([]*Metadata.ObjectMeta, 0)
				return request
			}
			request.ObjectMeta = append(request.ObjectMeta, &objectMeta)
		}
	}
	return request
}
