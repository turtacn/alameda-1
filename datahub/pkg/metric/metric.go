package metric

import (
	"time"

	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

// ContainerMetric Type/Kind alias
type ContainerMetricType = string
type ContainerMetricKind = string

// NodeMetric Type/Kind alias
type NodeMetricType = string
type NodeMetricKind = string

// GpuMetric Type/Kind alias
type GpuMetricType = string
type GpuMetricKind = string

const (
	// Container metric type definition
	TypeContainerCPUUsageSecondsPercentage ContainerMetricType = "cpu_usage_seconds_percentage"
	TypeContainerMemoryUsageBytes          ContainerMetricType = "memory_usage_bytes"

	// Node metric type definition
	TypeNodeCPUUsageSecondsPercentage NodeMetricType = "node_cpu_usage_seconds_percentage"
	TypeNodeMemoryTotalBytes          NodeMetricType = "node_memory_total_bytes"
	TypeNodeMemoryAvailableBytes      NodeMetricType = "node_memory_available_bytes"
	TypeNodeMemoryUsageBytes          NodeMetricType = "node_memory_usage_bytes"

	// GPU metric type definition
	TypeGpuDutyCycle                      GpuMetricType = "gpu_duty_cycle"
	TypeGpuDutyCycleLowerBound            GpuMetricType = "gpu_duty_cycle_lower_bound"
	TypeGpuDutyCycleUpperBound            GpuMetricType = "gpu_duty_cycle_upper_bound"
	TypeGpuMemoryUsedBytes                GpuMetricType = "gpu_memory_used_bytes"
	TypeGpuMemoryUsedBytesLowerBound      GpuMetricType = "gpu_memory_used_bytes_lower_bound"
	TypeGpuMemoryUsedBytesUpperBound      GpuMetricType = "gpu_memory_used_bytes_upper_bound"
	TypeGpuPowerUsageMilliWatts           GpuMetricType = "gpu_power_usage_milli_watts"
	TypeGpuPowerUsageMilliWattsLowerBound GpuMetricType = "gpu_power_usage_milli_watts_lower_bound"
	TypeGpuPowerUsageMilliWattsUpperBound GpuMetricType = "gpu_power_usage_milli_watts_upper_bound"
	TypeGpuTemperatureCelsius             GpuMetricType = "gpu_temperature_celsius"
	TypeGpuTemperatureCelsiusLowerBound   GpuMetricType = "gpu_temperature_celsius_lower_bound"
	TypeGpuTemperatureCelsiusUpperBound   GpuMetricType = "gpu_temperature_celsius_upper_bound"
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
	TypeToDatahubMetricType = map[string]datahub_v1alpha1.MetricType{
		TypeContainerCPUUsageSecondsPercentage: datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		TypeContainerMemoryUsageBytes:          datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES,

		TypeNodeCPUUsageSecondsPercentage: datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
		TypeNodeMemoryUsageBytes:          datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES,

		TypeGpuDutyCycle:                      datahub_v1alpha1.MetricType_DUTY_CYCLE,
		TypeGpuDutyCycleLowerBound:            datahub_v1alpha1.MetricType_DUTY_CYCLE,
		TypeGpuDutyCycleUpperBound:            datahub_v1alpha1.MetricType_DUTY_CYCLE,
		TypeGpuMemoryUsedBytes:                datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES,
		TypeGpuMemoryUsedBytesLowerBound:      datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES,
		TypeGpuMemoryUsedBytesUpperBound:      datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES,
		TypeGpuPowerUsageMilliWatts:           datahub_v1alpha1.MetricType_POWER_USAGE_WATTS,
		TypeGpuPowerUsageMilliWattsLowerBound: datahub_v1alpha1.MetricType_POWER_USAGE_WATTS,
		TypeGpuPowerUsageMilliWattsUpperBound: datahub_v1alpha1.MetricType_POWER_USAGE_WATTS,
		TypeGpuTemperatureCelsius:             datahub_v1alpha1.MetricType_TEMPERATURE_CELSIUS,
		TypeGpuTemperatureCelsiusLowerBound:   datahub_v1alpha1.MetricType_TEMPERATURE_CELSIUS,
		TypeGpuTemperatureCelsiusUpperBound:   datahub_v1alpha1.MetricType_TEMPERATURE_CELSIUS,
	}
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
