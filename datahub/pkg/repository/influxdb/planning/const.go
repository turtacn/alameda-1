package planning

import (
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

const (
	Container  influxdb.Measurement = "container"
	Controller influxdb.Measurement = "controller"
)
