package influxdb

import (
	//DaoRecommendationTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/recommendations/types"
	RepoInfluxRecommendation "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/recommendations"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiRecommendations "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/recommendations"
)

type NodeRecommendations struct {
	InfluxDBConfig InternalInflux.Config
}

func NewNodeRecommendationsWithConfig(config InternalInflux.Config) *NodeRecommendations {
	return &NodeRecommendations{InfluxDBConfig: config}
}

func (c *NodeRecommendations) CreateRecommendations(recommendations []*ApiRecommendations.NodeRecommendation) error {
	repository := RepoInfluxRecommendation.NewNodeRepository(&c.InfluxDBConfig)
	return repository.CreateRecommendations(recommendations)
}

func (c *NodeRecommendations) ListRecommendations(in *ApiRecommendations.ListNodeRecommendationsRequest) ([]*ApiRecommendations.NodeRecommendation, error) {
	repository := RepoInfluxRecommendation.NewNodeRepository(&c.InfluxDBConfig)
	return repository.ListRecommendations(in)
}
