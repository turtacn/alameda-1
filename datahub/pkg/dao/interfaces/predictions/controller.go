package predictions

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
)

func NewControllerPredictionsDAO(config config.Config) types.ControllerPredictionsDAO {
	return influxdb.NewControllerPredictionsWithConfig(*config.InfluxDB)
}
