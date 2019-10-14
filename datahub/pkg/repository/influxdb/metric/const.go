package metric

import (
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

const (
	ContainerCpu    InternalInflux.Measurement = "container_cpu"
	ContainerMemory InternalInflux.Measurement = "container_memory"
	NodeCpu         InternalInflux.Measurement = "node_cpu"
	NodeMemory      InternalInflux.Measurement = "node_memory"
)
