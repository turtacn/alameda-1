package predictions

import (
	"github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

const (
	Node      influxdb.Measurement = "node"
	Container influxdb.Measurement = "container"
)
