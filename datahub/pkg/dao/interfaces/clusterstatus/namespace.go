package clusterstatus

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
)

func NewNamespaceDAO(config config.Config) types.NamespaceDAO {
	return influxdb.NewNamespaceWithConfig(*config.InfluxDB)
}
