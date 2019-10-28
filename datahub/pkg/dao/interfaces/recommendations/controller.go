package recommendations

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/recommendations/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/recommendations/types"
)

func NewControllerRecommendationsDAO(config config.Config) types.ControllerRecommendationsDAO {
	return influxdb.NewControllerRecommendationsWithConfig(*config.InfluxDB)
}
