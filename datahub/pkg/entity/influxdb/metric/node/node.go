package node

type Field = string
type Tag = string

const (
	NodeTime   Tag = "time"
	Name       Tag = "name"
	MetricType Tag = "metric_type"

	Value Field = "value"
)

var (
	Tags   = []Tag{Name, MetricType}
	Fields = []Field{Value}

	MetricDatabaseName    = "alameda_metric"
	MetricMeasurementName = "node"
)
