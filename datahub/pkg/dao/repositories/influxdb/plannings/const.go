package plannings

import (
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

const (
	Container  influxdb.Measurement = "container"
	Controller influxdb.Measurement = "controller"
	App        influxdb.Measurement = "app"
	Namespace  influxdb.Measurement = "namespace"
	Node       influxdb.Measurement = "node"
	Cluster    influxdb.Measurement = "cluster"
)
