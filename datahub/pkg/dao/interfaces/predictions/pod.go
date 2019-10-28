package predictions

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
)

func NewPodPredictionsDAO(config config.Config) types.PodPredictionsDAO {
	return influxdb.NewPodPredictionsWithConfig(*config.InfluxDB)
}
