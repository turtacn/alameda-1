package plannings

import (
	"github.com/containers-ai/alameda/datahub/pkg/config"
	"github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/plannings/influxdb"
)

func NewContainerPlanningsDAO(config config.Config) *influxdb.ContainerPlannings {
	return influxdb.NewContainerPlanningsWithConfig(*config.InfluxDB)
}

func NewControllerPlanningsDAO(config config.Config) *influxdb.ControllerPlannings {
	return influxdb.NewControllerPlanningsWithConfig(*config.InfluxDB)
}

func NewAppPlanningsDAO(config config.Config) *influxdb.AppPlannings {
	return influxdb.NewAppPlanningsWithConfig(*config.InfluxDB)
}

func NewNamespacePlanningsDAO(config config.Config) *influxdb.NamespacePlannings {
	return influxdb.NewNamespacePlanningsWithConfig(*config.InfluxDB)
}

func NewNodePlanningsDAO(config config.Config) *influxdb.NodePlannings {
	return influxdb.NewNodePlanningsWithConfig(*config.InfluxDB)
}

func NewClusterPlanningsDAO(config config.Config) *influxdb.ClusterPlannings {
	return influxdb.NewClusterPlanningsWithConfig(*config.InfluxDB)
}
