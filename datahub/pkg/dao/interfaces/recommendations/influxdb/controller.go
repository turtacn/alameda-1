package influxdb

import (
	DaoRecommendationTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/recommendations/types"
	RepoInfluxRecommendation "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/recommendations"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiRecommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
)

type ControllerRecommendations struct {
	InfluxDBConfig InternalInflux.Config
}

func NewControllerRecommendationsWithConfig(config InternalInflux.Config) DaoRecommendationTypes.ControllerRecommendationsDAO {
	return &ControllerRecommendations{InfluxDBConfig: config}
}

func (c *ControllerRecommendations) AddControllerRecommendations(controllerRecommendations []*ApiRecommendations.ControllerRecommendation) error {
	controllerRepository := RepoInfluxRecommendation.NewControllerRepository(&c.InfluxDBConfig)
	return controllerRepository.CreateControllerRecommendations(controllerRecommendations)
}

func (c *ControllerRecommendations) ListControllerRecommendations(in *ApiRecommendations.ListControllerRecommendationsRequest) ([]*ApiRecommendations.ControllerRecommendation, error) {
	controllerRepository := RepoInfluxRecommendation.NewControllerRepository(&c.InfluxDBConfig)
	return controllerRepository.ListControllerRecommendations(in)
}
