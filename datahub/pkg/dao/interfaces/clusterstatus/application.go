package clusterstatus

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
)

func NewApplicationDAO(config config.Config) types.ApplicationDAO {
	return influxdb.NewApplicationWithConfig(*config.InfluxDB)
}
