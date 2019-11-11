package predictions

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
)

func NewClusterPredictionsDAO(config config.Config) types.ClusterPredictionsDAO {
	return influxdb.NewClusterPredictionsWithConfig(*config.InfluxDB)
}
