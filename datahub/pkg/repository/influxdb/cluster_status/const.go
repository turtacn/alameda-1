package clusterstatus

import (
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

const (
	// Node is node measurement
	Node influxdb.Measurement = "node"
	// Container is container measurement
	Container  influxdb.Measurement = "container"
	Controller influxdb.Measurement = "controller"
)
