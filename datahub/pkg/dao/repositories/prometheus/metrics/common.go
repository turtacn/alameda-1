package metrics

import (
	"github.com/containers-ai/alameda/pkg/utils/log"
)

var (
	scope = log.RegisterScope("metrics_prometheu", "metrics repository fetching from Prometheus", 0)
)

const (
	// Metric name to query from prometheus
	NodeMemoryBytesTotalMetricName = "node:node_memory_bytes_total:sum"
	// Label name in prometheus metric
	NodeMemoryBytesTotalLabelNode = "node"

	// Metric name to query from prometheus
	NodeMemoryBytesAvailableMetricName = "node:node_memory_bytes_available:sum"
	// Label name in prometheus metric
	NodeMemoryBytesAvailableLabelNode = "node"

	// Metric name to query from prometheus
	NodeMemoryUtilizationMetricName = "node:node_memory_utilisation_2:"
	// Label name in prometheus metric
	NodeMemoryUtilizationLabelNode = "node"

	// Label name in prometheus metric
	NodeMemoryBytesUsageLabelNode = "node"

	// Metric name to query from prometheus
	NodeCpuUsagePercentageMetricNameSum = "node:node_num_cpu:sum"
	NodeCpuUsagePercentageMetricNameAvg = "node:node_cpu_utilisation:avg1m"
	// Label name in prometheus metric
	NodeCpuUsagePercentageLabelNode = "node"

	// Metric name to query from prometheus
	ContainerCpuUsagePercentageMetricName = "namespace_pod_name_container_name:container_cpu_usage_seconds_total:sum_rate"
	// Label name in prometheus metric
	ContainerCpuUsagePercentageLabelNamespace     = "namespace"
	ContainerCpuUsagePercentageLabelPodName       = "pod_name"
	ContainerCpuUsagePercentageLabelContainerName = "container_name"

	// Metric name to query from prometheus
	ContainerMemoryUsageBytesMetricName = "container_memory_usage_bytes"
	// Label name in prometheus metric
	ContainerMemoryUsageBytesLabelNamespace     = "namespace"
	ContainerMemoryUsageBytesLabelPodName       = "pod_name"
	ContainerMemoryUsageBytesLabelContainerName = "container_name"
)
