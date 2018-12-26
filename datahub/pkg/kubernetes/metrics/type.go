package metrics

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
	LabelSelectorKeysAvailableForMetrics = map[MetricType]LabelSelectorKeys{
		MetricTypeContainerCPUUsageSecondsPercentage: LabelSelectorKeys{LabelSelectorKeyNamespace, LabelSelectorKeyPodName, LabelSelectorKeyContainerName},
		MetricTypeContainerMemoryUsageBytes:          LabelSelectorKeys{LabelSelectorKeyNamespace, LabelSelectorKeyPodName, LabelSelectorKeyContainerName},
		MetricTypeNodeCPUUsageSecondsPercentage:      LabelSelectorKeys{LabelSelectorKeyNodeName},
		MetricTypeNodeMemoryUsageBytes:               LabelSelectorKeys{LabelSelectorKeyNodeName},
	}
)

type MetricType int

type StringOperator int

type LabelSelectorKey string

type LabelSelectorKeys []LabelSelectorKey

func (lsks LabelSelectorKeys) Contains(keyToVerify LabelSelectorKey) bool {

	for _, keyExist := range lsks {
		if keyExist == keyToVerify {
			return true
		}
	}

	return false
}
