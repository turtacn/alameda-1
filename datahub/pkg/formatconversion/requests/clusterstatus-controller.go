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
		// Normalize request
		objectMeta := NewObjectMeta(ctl.GetObjectMeta())
		objectMeta.NodeName = ""

		controller := DaoClusterTypes.NewController()
		controller.ObjectMeta = objectMeta
		controller.Kind = ctl.GetKind().String()
		controller.Replicas = ctl.GetReplicas()
		controller.SpecReplicas = ctl.GetSpecReplicas()
		for _, owner := range ctl.GetOwnerReferences() {
			controller.OwnerReferences = append(controller.OwnerReferences, NewOwnerReference(owner))
		}
		controller.AlamedaControllerSpec = NewAlamedaControllerSpec(ctl.GetAlamedaControllerSpec())

		controllers = append(controllers, controller)
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
	request.Kind = r.GetKind().String()
	return request
}
