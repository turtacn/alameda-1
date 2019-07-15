package impl

import (
	RepoInfluxRecommendation "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb/recommendation"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

type Controller struct {
	InfluxDBConfig InternalInflux.Config
}

func (c *Controller) AddControllerRecommendations(controllerRecommendations []*datahub_v1alpha1.ControllerRecommendation) error {
	controllerRepository := RepoInfluxRecommendation.NewControllerRepository(&c.InfluxDBConfig)
	return controllerRepository.CreateControllerRecommendations(controllerRecommendations)
}

func (c *Controller) ListControllerRecommendations(in *datahub_v1alpha1.ListControllerRecommendationsRequest) ([]*datahub_v1alpha1.ControllerRecommendation, error) {
	controllerRepository := RepoInfluxRecommendation.NewControllerRepository(&c.InfluxDBConfig)
	return controllerRepository.ListControllerRecommendations(in)
}
