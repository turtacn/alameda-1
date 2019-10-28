package predictions

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
)

func NewNodePredictionsDAO(config config.Config) types.NodePredictionsDAO {
	return influxdb.NewNodePredictionsWithConfig(*config.InfluxDB)
}
