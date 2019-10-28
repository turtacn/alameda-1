package clusterstatus

import (
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type ControllerOperation interface {
	CreateControllers([]*ApiResources.Controller) error
	ListControllers(*ApiResources.ListControllersRequest) ([]*ApiResources.Controller, error)
	DeleteControllers(*ApiResources.DeleteControllersRequest) error
}
