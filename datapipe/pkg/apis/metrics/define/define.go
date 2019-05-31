package define

import (
	"time"

	//datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	//dataPipeMetricsAPI "github.com/containers-ai/api/datapipe/metrics"
	commonAPI "github.com/containers-ai/api/common"
	datahubMetricsAPI "github.com/containers-ai/api/datahub/metrics"
)

// ContainerMetricType Type alias
type ContainerMetricType = string

// ContainerMetricKind Kind alias
type ContainerMetricKind = string

// NodeMetricType Type alias
type NodeMetricType = string

type NodeMetricKind = string

const (
	// TypeContainerCPUUsageSecondsPercentage Metric type of container cpu usage
	TypeContainerCPUUsageSecondsPercentage ContainerMetricType = "cpu_usage_seconds_percentage"
	// TypeContainerMemoryUsageBytes Metric type of container memory usage
	TypeContainerMemoryUsageBytes ContainerMetricType = "memory_usage_bytes"

	// TypeNodeCPUUsageSecondsPercentage Metric type of cpu usage
	TypeNodeCPUUsageSecondsPercentage NodeMetricType = "node_cpu_usage_seconds_percentage"
	// TypeNodeMemoryTotalBytes Metric type of memory total
	TypeNodeMemoryTotalBytes NodeMetricType = "node_memory_total_bytes"
	// TypeNodeMemoryAvailableBytes Metric type of memory available
	TypeNodeMemoryAvailableBytes NodeMetricType = "node_memory_available_bytes"
	// TypeNodeMemoryUsageBytes Metric type of memory usage
	TypeNodeMemoryUsageBytes NodeMetricType = "node_memory_usage_bytes"
)

const (
	ContainerMetricKindRaw        ContainerMetricKind = "raw"
	ContainerMetricKindUpperbound ContainerMetricKind = "upper_bound"
	ContainerMetricKindLowerbound ContainerMetricKind = "lower_bound"

	NodeMetricKindRaw        NodeMetricKind = "raw"
	NodeMetricKindUpperbound NodeMetricKind = "upper_bound"
	NodeMetricKindLowerbound NodeMetricKind = "lower_bound"
)

var (
	// TypeToDatahubMetricType Type to datahub metric type
	TypeToDatahubMetricType = map[string]datahubMetricsAPI.MetricType{
		TypeContainerCPUUsageSecondsPercentage: datahubMetricsAPI.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		TypeContainerMemoryUsageBytes:          datahubMetricsAPI.MetricType_MEMORY_USAGE_BYTES,
		TypeNodeCPUUsageSecondsPercentage:      datahubMetricsAPI.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		TypeNodeMemoryUsageBytes:               datahubMetricsAPI.MetricType_MEMORY_USAGE_BYTES,
	}
)

var (
	MetricDatabaseName    = ""
	MetricMeasurementName = ""
	MetricColumns         = []string{
		"pod_namespace",
		"pod_name",
		"name",
		"metric_type",
		"value"}

	MetricColumnsTypes = []commonAPI.ColumnType{
		commonAPI.ColumnType_COLUMNTYPE_TAG,
		commonAPI.ColumnType_COLUMNTYPE_TAG,
		commonAPI.ColumnType_COLUMNTYPE_TAG,
		commonAPI.ColumnType_COLUMNTYPE_TAG,
		commonAPI.ColumnType_COLUMNTYPE_FIELD}

	MetricDataTypes = []commonAPI.DataType{
		commonAPI.DataType_DATATYPE_STRING,
		commonAPI.DataType_DATATYPE_STRING,
		commonAPI.DataType_DATATYPE_STRING,
		commonAPI.DataType_DATATYPE_STRING,
		commonAPI.DataType_DATATYPE_FLOAT64}
)

// Sample Data struct representing timestamp and metric value of metric data point
type Sample struct {
	Timestamp time.Time
	Value     string
}

type SamplesByAscTimestamp []Sample

func (d SamplesByAscTimestamp) Len() int {
	return len(d)
}
func (d SamplesByAscTimestamp) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
func (d SamplesByAscTimestamp) Less(i, j int) bool {
	return d[i].Timestamp.Unix() < d[j].Timestamp.Unix()
}

type SamplesByDescTimestamp []Sample

func (d SamplesByDescTimestamp) Len() int {
	return len(d)
}
func (d SamplesByDescTimestamp) Swap(i, j int) {
	d[i], d[j] = d[j], d[i]
}
func (d SamplesByDescTimestamp) Less(i, j int) bool {
	return d[i].Timestamp.Unix() > d[j].Timestamp.Unix()
}
