package container

type Field = string
type Tag = string

const (
	PodTime      Tag = "time"
	PodNamespace Tag = "pod_namespace"
	PodName      Tag = "pod_name"
	Name         Tag = "name"
	MetricType   Tag = "metric_type"

	Value Field = "value"
)

var (
	Tags   = []Tag{PodNamespace, PodName, Name, MetricType}
	Fields = []Field{Value}

	MetricDatabaseName    = "alameda_metric"
	MetricMeasurementName = "container"
	MetricColumns         = []string{
		PodNamespace,
		PodName,
		Name,
		MetricType,
		Value}
)
