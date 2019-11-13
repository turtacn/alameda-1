package influxdb

import (
	RepoInfluxRecommendation "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/recommendations"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	ApiRecommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
)

var (
	scope = Log.RegisterScope("recommendation_dao_implement", "recommended dao implement", 0)
)

// Container Implements ContainerOperation interface
type ContainerRecommendations struct {
	InfluxDBConfig InternalInflux.Config
}

func NewContainerRecommendationsWithConfig(config InternalInflux.Config) *ContainerRecommendations {
	return &ContainerRecommendations{InfluxDBConfig: config}
}

// AddPodRecommendations add pod recommendations to database
func (c *ContainerRecommendations) CreatePodRecommendations(in *ApiRecommendations.CreatePodRecommendationsRequest) error {
	containerRepository := RepoInfluxRecommendation.NewContainerRepository(&c.InfluxDBConfig)
	return containerRepository.CreateContainerRecommendations(in)
}

// ListPodRecommendations list pod recommendations
func (c *ContainerRecommendations) ListPodRecommendations(in *ApiRecommendations.ListPodRecommendationsRequest) ([]*ApiRecommendations.PodRecommendation, error) {
	containerRepository := RepoInfluxRecommendation.NewContainerRepository(&c.InfluxDBConfig)
	return containerRepository.ListContainerRecommendations(in)
}

func (c *ContainerRecommendations) ListAvailablePodRecommendations(in *ApiRecommendations.ListPodRecommendationsRequest) ([]*ApiRecommendations.PodRecommendation, error) {
	containerRepository := RepoInfluxRecommendation.NewContainerRepository(&c.InfluxDBConfig)
	return containerRepository.ListAvailablePodRecommendations(in)
}
