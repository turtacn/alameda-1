package impl

import (
	influxdb_repository "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"

	influxdb_repository_recommendation "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb/recommendation"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

var (
	scope = log.RegisterScope("recommendation_dao_implement", "recommended dao implement", 0)
)

// Container Implements ContainerOperation interface
type Container struct {
	InfluxDBConfig influxdb_repository.Config
}

// AddPodRecommendations add pod recommendations to database
func (container *Container) AddPodRecommendations(podRecommendations []*datahub_v1alpha1.PodRecommendation) error {
	containerRepository := influxdb_repository_recommendation.NewContainerRepository(&container.InfluxDBConfig)
	return containerRepository.CreateContainerRecommendations(podRecommendations)
}

// ListPodRecommendations list pod recommendations
func (container *Container) ListPodRecommendations(podNamespacedName *datahub_v1alpha1.NamespacedName,
	queryCondition *datahub_v1alpha1.QueryCondition,
	kind datahub_v1alpha1.Kind) ([]*datahub_v1alpha1.PodRecommendation, error) {
	containerRepository := influxdb_repository_recommendation.NewContainerRepository(&container.InfluxDBConfig)
	return containerRepository.ListContainerRecommendations(podNamespacedName, queryCondition, kind)
}
