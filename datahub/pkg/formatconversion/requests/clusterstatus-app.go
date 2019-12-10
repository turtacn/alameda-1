package requests

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type CreateApplicationsRequestExtended struct {
	ApiResources.CreateApplicationsRequest
}

type ListApplicationsRequestExtended struct {
	*ApiResources.ListApplicationsRequest
}

type DeleteApplicationsRequestExtended struct {
	*ApiResources.DeleteApplicationsRequest
}

func NewApplication(application *ApiResources.Application) *DaoClusterTypes.Application {
	if application != nil {
		// Normalize request
		objectMeta := NewObjectMeta(application.GetObjectMeta())
		objectMeta.NodeName = ""

		app := DaoClusterTypes.Application{}
		app.ObjectMeta = &objectMeta
		app.AlamedaApplicationSpec = NewAlamedaApplicationSpec(application.GetAlamedaApplicationSpec())
		app.Controllers = make([]*DaoClusterTypes.Controller, 0)

		return &app
	}
	return nil
}

func (p *CreateApplicationsRequestExtended) Validate() error {
	return nil
}

func (p *CreateApplicationsRequestExtended) ProduceApplications() []*DaoClusterTypes.Application {
	applications := make([]*DaoClusterTypes.Application, 0)

	for _, app := range p.GetApplications() {
		applications = append(applications, NewApplication(app))
	}

	return applications
}

func (p *ListApplicationsRequestExtended) Validate() error {
	return nil
}

func (p *ListApplicationsRequestExtended) ProduceRequest() *DaoClusterTypes.ListApplicationsRequest {
	request := DaoClusterTypes.NewListApplicationsRequest()
	if p.GetObjectMeta() != nil {
		for _, meta := range p.GetObjectMeta() {
			// Normalize request
			objectMeta := NewObjectMeta(meta)
			objectMeta.NodeName = ""

			if objectMeta.IsEmpty() {
				request.ApplicationObjectMeta = make([]*DaoClusterTypes.ApplicationObjectMeta, 0)
				return request
			}
			request.ApplicationObjectMeta = append(request.ApplicationObjectMeta, DaoClusterTypes.NewApplicationObjectMeta(&objectMeta, ""))
		}
	}
	return request
}

func (p *DeleteApplicationsRequestExtended) Validate() error {
	return nil
}

func (p *DeleteApplicationsRequestExtended) ProduceRequest() *DaoClusterTypes.DeleteApplicationsRequest {
	request := DaoClusterTypes.NewDeleteApplicationsRequest()
	if p.GetObjectMeta() != nil {
		for _, meta := range p.GetObjectMeta() {
			// Normalize request
			objectMeta := NewObjectMeta(meta)
			objectMeta.NodeName = ""

			if objectMeta.IsEmpty() {
				request.ApplicationObjectMeta = make([]*DaoClusterTypes.ApplicationObjectMeta, 0)
				return request
			}
			request.ApplicationObjectMeta = append(request.ApplicationObjectMeta, DaoClusterTypes.NewApplicationObjectMeta(&objectMeta, ""))
		}
	}
	return request
}
