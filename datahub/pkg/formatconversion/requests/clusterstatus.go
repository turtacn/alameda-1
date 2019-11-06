package requests

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type CreateApplicationsRequestExtended struct {
	ApiResources.CreateApplicationsRequest
}

func (r *CreateApplicationsRequestExtended) Validate() error {
	return nil
}

func (r *CreateApplicationsRequestExtended) ProduceApplications() []*DaoClusterTypes.Application {
	applications := make([]*DaoClusterTypes.Application, 0)

	for _, app := range r.GetApplications() {
		// Normalize request
		objectMeta := NewObjectMeta(app.GetObjectMeta())
		objectMeta.NodeName = ""

		application := DaoClusterTypes.NewApplication()
		application.ObjectMeta = objectMeta
		applications = append(applications, application)
	}

	return applications
}

type CreateNamespacesRequestExtended struct {
	ApiResources.CreateNamespacesRequest
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

type ListApplicationsRequestExtended struct {
	*ApiResources.ListApplicationsRequest
}

func (r *ListApplicationsRequestExtended) Validate() error {
	return nil
}

func (r *ListApplicationsRequestExtended) ProduceRequest() DaoClusterTypes.ListApplicationsRequest {
	request := DaoClusterTypes.NewListApplicationsRequest()
	if r.GetObjectMeta() != nil {
		for _, meta := range r.GetObjectMeta() {
			// Normalize request
			objectMeta := NewObjectMeta(meta)
			objectMeta.NodeName = ""

			if objectMeta.IsEmpty() {
				return DaoClusterTypes.NewListApplicationsRequest()
			}
			request.ObjectMeta = append(request.ObjectMeta, objectMeta)
		}
	}
	return request
}

type ListNamespacesRequestExtended struct {
	*ApiResources.ListNamespacesRequest
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
