package metrics

import (
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

const (
	ContainerCpu      InternalInflux.Measurement = "container_cpu"
	ContainerMemory   InternalInflux.Measurement = "container_memory"
	NodeCpu           InternalInflux.Measurement = "node_cpu"
	NodeMemory        InternalInflux.Measurement = "node_memory"
	ApplicationCpu    InternalInflux.Measurement = "application_cpu"
	ApplicationMemory InternalInflux.Measurement = "application_memory"
	ClusterCpu        InternalInflux.Measurement = "cluster_cpu"
	ClusterMemory     InternalInflux.Measurement = "cluster_memory"
	NamespaceCpu      InternalInflux.Measurement = "namespace_cpu"
	NamespaceMemory   InternalInflux.Measurement = "namespace_memory"
	ControllerCpu     InternalInflux.Measurement = "controller_cpu"
	ControllerMemory  InternalInflux.Measurement = "controller_memory"
)
