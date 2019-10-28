package types

import (
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
)

// ContainerOperation defines container measurement operation of recommendation database
type ControllerPlanningsDAO interface {
	AddControllerPlannings([]*ApiPlannings.ControllerPlanning) error
	ListControllerPlannings(in *ApiPlannings.ListControllerPlanningsRequest) ([]*ApiPlannings.ControllerPlanning, error)
}
