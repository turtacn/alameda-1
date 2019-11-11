package predictions

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
)

func NewApplicationPredictionsDAO(config config.Config) types.ApplicationPredictionsDAO {
	return influxdb.NewApplicationPredictionsWithConfig(*config.InfluxDB)
}
