package requests

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type CreateControllersRequestExtended struct {
	ApiResources.CreateControllersRequest
}

type ListControllersRequestExtended struct {
	*ApiResources.ListControllersRequest
}

func (r *CreateControllersRequestExtended) Validate() error {
	return nil
}

func (r *CreateControllersRequestExtended) ProduceControllers() []*DaoClusterTypes.Controller {
	controllers := make([]*DaoClusterTypes.Controller, 0)

	for _, ctl := range r.GetControllers() {
		controllers = append(controllers, NewController(ctl))
	}

	return controllers
}

func (r *ListControllersRequestExtended) Validate() error {
	return nil
}

func (r *ListControllersRequestExtended) ProduceRequest() DaoClusterTypes.ListControllersRequest {
	request := DaoClusterTypes.NewListControllersRequest()
	if r.GetObjectMeta() != nil {
		for _, meta := range r.GetObjectMeta() {
			// Normalize request
			objectMeta := NewObjectMeta(meta)
			objectMeta.NodeName = ""

			if objectMeta.IsEmpty() {
				request := DaoClusterTypes.NewListControllersRequest()
				request.Kind = r.GetKind().String()
				return request
			}
			request.ObjectMeta = append(request.ObjectMeta, objectMeta)
		}
	}
	if r.GetKind() != ApiResources.Kind_KIND_UNDEFINED {
		request.Kind = r.GetKind().String()
	}
	return request
}

func NewController(controller *ApiResources.Controller) *DaoClusterTypes.Controller {
	if controller != nil {
		// Normalize request
		objectMeta := NewObjectMeta(controller.GetObjectMeta())
		objectMeta.NodeName = ""

		ctl := DaoClusterTypes.NewController()
		ctl.ObjectMeta = objectMeta
		ctl.Kind = controller.GetKind().String()
		ctl.Replicas = controller.GetReplicas()
		ctl.SpecReplicas = controller.GetSpecReplicas()
		ctl.AlamedaControllerSpec = NewAlamedaControllerSpec(controller.GetAlamedaControllerSpec())

		return ctl
	}
	return nil
}
