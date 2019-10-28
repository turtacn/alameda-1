package plannings

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/plannings/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/plannings/types"
)

func NewContainerPlanningsDAO(config config.Config) types.ContainerPlanningsDAO {
	return influxdb.NewContainerPlanningsWithConfig(*config.InfluxDB)
}
