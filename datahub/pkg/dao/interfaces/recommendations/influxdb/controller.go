package influxdb

import (
	RepoInfluxRecommendation "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/recommendations"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiRecommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
)

type ControllerRecommendations struct {
	InfluxDBConfig InternalInflux.Config
}

func NewControllerRecommendationsWithConfig(config InternalInflux.Config) *ControllerRecommendations {
	return &ControllerRecommendations{InfluxDBConfig: config}
}

func (c *ControllerRecommendations) CreateControllerRecommendations(controllerRecommendations []*ApiRecommendations.ControllerRecommendation) error {
	controllerRepository := RepoInfluxRecommendation.NewControllerRepository(&c.InfluxDBConfig)
	return controllerRepository.CreateControllerRecommendations(controllerRecommendations)
}

func (c *ControllerRecommendations) ListControllerRecommendations(in *ApiRecommendations.ListControllerRecommendationsRequest) ([]*ApiRecommendations.ControllerRecommendation, error) {
	controllerRepository := RepoInfluxRecommendation.NewControllerRepository(&c.InfluxDBConfig)
	return controllerRepository.ListControllerRecommendations(in)
}
