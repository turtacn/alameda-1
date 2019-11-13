package influxdb

import (
	//DaoRecommendationTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/recommendations/types"
	RepoInfluxRecommendation "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/recommendations"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiRecommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
)

type AppRecommendations struct {
	InfluxDBConfig InternalInflux.Config
}

func NewAppRecommendationsWithConfig(config InternalInflux.Config) *AppRecommendations {
	return &AppRecommendations{InfluxDBConfig: config}
}

func (c *AppRecommendations) CreateRecommendations(recommendations []*ApiRecommendations.ApplicationRecommendation) error {
	repository := RepoInfluxRecommendation.NewAppRepository(&c.InfluxDBConfig)
	return repository.CreateRecommendations(recommendations)
}

func (c *AppRecommendations) ListRecommendations(in *ApiRecommendations.ListApplicationRecommendationsRequest) ([]*ApiRecommendations.ApplicationRecommendation, error) {
	repository := RepoInfluxRecommendation.NewAppRepository(&c.InfluxDBConfig)
	return repository.ListRecommendations(in)
}
