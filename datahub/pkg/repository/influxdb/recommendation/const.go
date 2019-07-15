package recommendation

import (
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

const (
	// Container is container measurement
	Container  influxdb.Measurement = "container"
	Controller influxdb.Measurement = "controller"
)
