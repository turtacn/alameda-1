package clusterstatus

import (
	"github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
)

const (
	Node      influxdb.Measurement = "node"
	Container influxdb.Measurement = "container"
)
