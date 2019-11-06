package types

import (
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type ControllerDAO interface {
	CreateControllers([]*resources.Controller) error
	ListControllers(*resources.ListControllersRequest) ([]*resources.Controller, error)
	DeleteControllers(*resources.DeleteControllersRequest) error
}
