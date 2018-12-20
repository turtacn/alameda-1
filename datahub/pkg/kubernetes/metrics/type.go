package metrics

type MetricType int
type StringOperator int
type LabelSelectorKey string

const (
	MetricTypeContainerCPUUsageSecondsPercentage MetricType = 0
	MetricTypeContainerMemoryUsageBytes          MetricType = 1
	MetricTypeNodeCPUUsageSecondsPercentage      MetricType = 2
	MetricTypeNodeMemoryUsageBytes               MetricType = 3

	StringOperatorEqueal    StringOperator = 0
	StringOperatorNotEqueal StringOperator = 1

	LabelSelectorKeyNamespace     LabelSelectorKey = "namespace"
	LabelSelectorKeyPodName       LabelSelectorKey = "pod_name"
	LabelSelectorKeyContainerName LabelSelectorKey = "container_name"
	LabelSelectorKeyNodeName      LabelSelectorKey = "node_name"
)

var (
	LabelSelectorKeysAvailableForMetrics = map[MetricType][]LabelSelectorKey{
		MetricTypeContainerCPUUsageSecondsPercentage: []LabelSelectorKey{LabelSelectorKeyNamespace, LabelSelectorKeyPodName, LabelSelectorKeyContainerName},
		MetricTypeContainerMemoryUsageBytes:          []LabelSelectorKey{LabelSelectorKeyNamespace, LabelSelectorKeyPodName, LabelSelectorKeyContainerName},
		MetricTypeNodeCPUUsageSecondsPercentage:      []LabelSelectorKey{LabelSelectorKeyNodeName},
		MetricTypeNodeMemoryUsageBytes:               []LabelSelectorKey{LabelSelectorKeyNodeName},
	}
)
