package recommendations

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/recommendations/influxdb"
)

func NewContainerRecommendationsDAO(config config.Config) *influxdb.ContainerRecommendations {
	return influxdb.NewContainerRecommendationsWithConfig(*config.InfluxDB)
}

func NewControllerRecommendationsDAO(config config.Config) *influxdb.ControllerRecommendations {
	return influxdb.NewControllerRecommendationsWithConfig(*config.InfluxDB)
}

func NewAppRecommendationsDAO(config config.Config) *influxdb.AppRecommendations {
	return influxdb.NewAppRecommendationsWithConfig(*config.InfluxDB)
}

func NewNamespaceRecommendationsDAO(config config.Config) *influxdb.NamespaceRecommendations {
	return influxdb.NewNamespaceRecommendationsWithConfig(*config.InfluxDB)
}

func NewNodeRecommendationsDAO(config config.Config) *influxdb.NodeRecommendations {
	return influxdb.NewNodeRecommendationsWithConfig(*config.InfluxDB)
}

func NewClusterRecommendationsDAO(config config.Config) *influxdb.ClusterRecommendations {
	return influxdb.NewClusterRecommendationsWithConfig(*config.InfluxDB)
}
