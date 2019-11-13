package clusterstatus

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
)

func NewControllerDAO(config config.Config) types.ControllerDAO {
	return influxdb.NewControllerWithConfig(*config.InfluxDB)
}
