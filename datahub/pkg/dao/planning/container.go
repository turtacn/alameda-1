package planning

import (
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

// ContainerOperation defines container measurement operation of recommendation database
type ContainerOperation interface {
	AddPodPlannings(in *DatahubV1alpha1.CreatePodPlanningsRequest) error
	ListPodPlannings(in *DatahubV1alpha1.ListPodPlanningsRequest) ([]*DatahubV1alpha1.PodPlanning, error)
}
