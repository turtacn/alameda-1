package metrics

import (
	"time"
)

type MetricsDB interface {
	Connect() error
	Close() error
	Query(q Query) (QueryResponse, error)
	ListContainerCPUUsageSecondsPercentage(q Query) (QueryResponse, error)
	ListContainerMemoryUsageBytes(q Query) (QueryResponse, error)
}

type Query struct {
	Metric         MetricType
	TimeSelector   TimeSelector
	LabelSelectors []LabelSelector
}

type QueryResponse struct {
	Metric  MetricType
	Results []Data
}

type Data struct {
	Labels  map[string]string
	Samples []Sample
}

type Sample struct {
	Time  time.Time
	Value float64
}

type TimeSelector interface {
	is_timeselector()
}

type Timestamp struct {
	T time.Time
}

func (t *Timestamp) is_timeselector() {}

// TimeRange Represent the range of data to query.
// Field Step represent the time width between each sample.
type TimeRange struct {
	StartTime, EndTime time.Time
	Step               time.Duration
}

func (t *TimeRange) is_timeselector() {}

type Since struct {
	Duration time.Duration
}

func (t *Since) is_timeselector() {}

type LabelSelector struct {
	Key, Value string
	Op         StringOperator
}

type BackendQueryRequestFactory interface {
	BuildServiceRequest() (interface{}, error)
}
