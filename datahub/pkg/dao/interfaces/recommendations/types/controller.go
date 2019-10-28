package types

import (
	ApiRecommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
)

// ContainerOperation defines container measurement operation of recommendation database
type ControllerRecommendationsDAO interface {
	AddControllerRecommendations([]*ApiRecommendations.ControllerRecommendation) error
	ListControllerRecommendations(in *ApiRecommendations.ListControllerRecommendationsRequest) ([]*ApiRecommendations.ControllerRecommendation, error)
}
