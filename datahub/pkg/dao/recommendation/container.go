package recommendation

import (
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

// ContainerOperation defines container measurement operation of recommendation database
type ContainerOperation interface {
	AddPodRecommendations([]*datahub_v1alpha1.PodRecommendation) error
	ListPodRecommendations(podNamespacedName *datahub_v1alpha1.NamespacedName, queryCondition *datahub_v1alpha1.QueryCondition) ([]*datahub_v1alpha1.PodRecommendation, error)
}
