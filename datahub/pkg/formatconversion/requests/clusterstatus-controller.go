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

type DeleteControllersRequestExtended struct {
	*ApiResources.DeleteControllersRequest
}

func NewController(controller *ApiResources.Controller) *DaoClusterTypes.Controller {
	if controller != nil {
		// Normalize request
		objectMeta := NewObjectMeta(controller.GetObjectMeta())
		objectMeta.NodeName = ""

		ctl := DaoClusterTypes.Controller{}
		ctl.ObjectMeta = &objectMeta
		ctl.Kind = controller.GetKind().String()
		ctl.Replicas = controller.GetReplicas()
		ctl.SpecReplicas = controller.GetSpecReplicas()
		ctl.AlamedaControllerSpec = NewAlamedaControllerSpec(controller.GetAlamedaControllerSpec())

		return &ctl
	}
	return nil
}

func (p *CreateControllersRequestExtended) Validate() error {
	return nil
}

func (p *CreateControllersRequestExtended) ProduceControllers() []*DaoClusterTypes.Controller {
	controllers := make([]*DaoClusterTypes.Controller, 0)

	for _, ctl := range p.GetControllers() {
		controllers = append(controllers, NewController(ctl))
	}

	return controllers
}

func (p *ListControllersRequestExtended) Validate() error {
	return nil
}

func (p *ListControllersRequestExtended) ProduceRequest() *DaoClusterTypes.ListControllersRequest {
	request := DaoClusterTypes.NewListControllersRequest()

	if p.GetObjectMeta() != nil {
		for _, meta := range p.GetObjectMeta() {
			// Normalize request
			objectMeta := NewObjectMeta(meta)
			objectMeta.NodeName = ""

			if objectMeta.IsEmpty() {
				controllerObjectMeta := DaoClusterTypes.NewControllerObjectMeta(nil, nil, p.GetKind().String(), "")
				request.ControllerObjectMeta = make([]*DaoClusterTypes.ControllerObjectMeta, 0)
				request.ControllerObjectMeta = append(request.ControllerObjectMeta, controllerObjectMeta)
				return request
			}

			controllerObjectMeta := DaoClusterTypes.NewControllerObjectMeta(&objectMeta, nil, p.GetKind().String(), "")
			request.ControllerObjectMeta = append(request.ControllerObjectMeta, controllerObjectMeta)
		}
	}

	if len(request.ControllerObjectMeta) == 0 {
		controllerObjectMeta := DaoClusterTypes.NewControllerObjectMeta(nil, nil, p.GetKind().String(), "")
		request.ControllerObjectMeta = append(request.ControllerObjectMeta, controllerObjectMeta)
	}

	return request
}

func (p *DeleteControllersRequestExtended) Validate() error {
	return nil
}

func (p *DeleteControllersRequestExtended) ProduceRequest() *DaoClusterTypes.DeleteControllersRequest {
	request := DaoClusterTypes.NewDeleteControllersRequest()

	if p.GetObjectMeta() != nil {
		for _, meta := range p.GetObjectMeta() {
			// Normalize request
			objectMeta := NewObjectMeta(meta)
			objectMeta.NodeName = ""

			if objectMeta.IsEmpty() {
				controllerObjectMeta := DaoClusterTypes.NewControllerObjectMeta(nil, nil, p.GetKind().String(), "")
				request.ControllerObjectMeta = make([]*DaoClusterTypes.ControllerObjectMeta, 0)
				request.ControllerObjectMeta = append(request.ControllerObjectMeta, controllerObjectMeta)
				return request
			}

			controllerObjectMeta := DaoClusterTypes.NewControllerObjectMeta(&objectMeta, nil, p.GetKind().String(), "")
			request.ControllerObjectMeta = append(request.ControllerObjectMeta, controllerObjectMeta)
		}
	}

	if len(request.ControllerObjectMeta) == 0 {
		controllerObjectMeta := DaoClusterTypes.NewControllerObjectMeta(nil, nil, p.GetKind().String(), "")
		request.ControllerObjectMeta = append(request.ControllerObjectMeta, controllerObjectMeta)
	}

	return request
}
