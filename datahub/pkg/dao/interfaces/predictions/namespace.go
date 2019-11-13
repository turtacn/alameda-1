package predictions

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
)

func NewNamespacePredictionsDAO(config config.Config) types.NamespacePredictionsDAO {
	return influxdb.NewNamespacePredictionsWithConfig(*config.InfluxDB)
}
