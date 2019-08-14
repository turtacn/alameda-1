package planning

import (
	"github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
)

const (
	Container  influxdb.Measurement = "container"
	Controller influxdb.Measurement = "controller"
)
