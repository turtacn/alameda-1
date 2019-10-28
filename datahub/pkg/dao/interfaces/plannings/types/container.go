package types

import (
	ApiPlannings "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/plannings"
)

// ContainerOperation defines container measurement operation of recommendation database
type ContainerPlanningsDAO interface {
	AddPodPlannings(in *ApiPlannings.CreatePodPlanningsRequest) error
	ListPodPlannings(in *ApiPlannings.ListPodPlanningsRequest) ([]*ApiPlannings.PodPlanning, error)
}
