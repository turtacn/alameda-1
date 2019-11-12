package clusterstatus

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
)

func NewPodDAO(config config.Config) types.PodDAO {
	return influxdb.NewPodWithConfig(*config.InfluxDB)
}
