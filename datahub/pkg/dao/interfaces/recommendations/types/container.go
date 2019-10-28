package types

import (
	ApiRecommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
)

// ContainerOperation defines container measurement operation of recommendation database
type ContainerRecommendationsDAO interface {
	AddPodRecommendations(in *ApiRecommendations.CreatePodRecommendationsRequest) error
	ListPodRecommendations(in *ApiRecommendations.ListPodRecommendationsRequest) ([]*ApiRecommendations.PodRecommendation, error)
	ListAvailablePodRecommendations(*ApiRecommendations.ListPodRecommendationsRequest) ([]*ApiRecommendations.PodRecommendation, error)
}
