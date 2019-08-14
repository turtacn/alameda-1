package planning

import (
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

// ContainerOperation defines container measurement operation of recommendation database
type ControllerOperation interface {
	AddControllerPlannings([]*DatahubV1alpha1.ControllerPlanning) error
	ListControllerPlannings(controllerNamespacedName *DatahubV1alpha1.NamespacedName, queryCondition *DatahubV1alpha1.QueryCondition) ([]*DatahubV1alpha1.ControllerPlanning, error)
}
