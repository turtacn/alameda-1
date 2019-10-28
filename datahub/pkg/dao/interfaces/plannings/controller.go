package plannings

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/plannings/influxdb"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/plannings/types"
)

func NewControllerPlanningsDAO(config config.Config) types.ControllerPlanningsDAO {
	return influxdb.NewControllerPlanningsWithConfig(*config.InfluxDB)
}
