package metrics

type MetricType int
type StringOperator int
type LabelSelectorKey string

const (
	MetricTypeContainerCPUUsageTotal     MetricType = 0
	MetricTypeContainerCPUUsageTotalRate MetricType = 1
	MetricTypeContainerMemoryUsage       MetricType = 2
	MetricTypeNodeCPUUsageSecondsAvg1M   MetricType = 3
	MetricTypeNodeMemoryUsageBytes       MetricType = 4

	StringOperatorEqueal    StringOperator = 0
	StringOperatorNotEqueal StringOperator = 1

	LabelSelectorKeyNamespace     LabelSelectorKey = "namespace"
	LabelSelectorKeyPodName       LabelSelectorKey = "pod_name"
	LabelSelectorKeyContainerName LabelSelectorKey = "container_name"
	LabelSelectorKeyNodeName      LabelSelectorKey = "node_name"
)
