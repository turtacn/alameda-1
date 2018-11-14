package metrics

type MetricType int
type StringOperator int

const (
	MetricTypeContainerCPUUsageTotal     MetricType = 0
	MetricTypeContainerCPUUsageTotalRate MetricType = 1
	MetricTypeContainerMemoryUsage       MetricType = 2

	StringOperatorEqueal    StringOperator = 0
	StringOperatorNotEqueal StringOperator = 1
)
