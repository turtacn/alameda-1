package recommendations

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/recommendations/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/recommendations/types"
)

func NewContainerRecommendationsDAO(config config.Config) types.ContainerRecommendationsDAO {
	return influxdb.NewContainerRecommendationsWithConfig(*config.InfluxDB)
}
