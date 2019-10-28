package enumconv

import (
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
)

// Metric Type/Kind alias
type MetricType = string
type MetricKind = string

// GpuMetric Type/Kind alias
type GpuMetricType = string
type GpuMetricKind = string

const (
	MetricTypeCPUUsageSecondsPercentage MetricType = "cpu_usage_seconds_percentage"
	MetricTypeMemoryUsageBytes          MetricType = "memory_usage_bytes"
	MetricTypePowerUsageWatts           MetricType = "power_usage_watts"
	MetricTypeTemperatureCelsius        MetricType = "temperature_celsius"
	MetricTypeDutyCycle                 MetricType = "duty_cycle"
	MetricTypeMemoryTotalBytes          MetricType = "memory_total_bytes"
	MetricTypeMemoryAvailableBytes      MetricType = "memory_available_bytes"

	MetricKindRaw        MetricKind = "raw"
	MetricKindUpperBound MetricKind = "upper_bound"
	MetricKindLowerBound MetricKind = "lower_bound"
)

// GPU metric type definition
const (
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

// TypeToDatahubMetricType Type to datahub metric type
var TypeToDatahubMetricType map[string]ApiCommon.MetricType = map[string]ApiCommon.MetricType{
	MetricTypeCPUUsageSecondsPercentage: ApiCommon.MetricType_CPU_USAGE_SECONDS_PERCENTAGE,
	MetricTypeMemoryUsageBytes:          ApiCommon.MetricType_MEMORY_USAGE_BYTES,
	MetricTypePowerUsageWatts:           ApiCommon.MetricType_POWER_USAGE_WATTS,
	MetricTypeTemperatureCelsius:        ApiCommon.MetricType_TEMPERATURE_CELSIUS,
	MetricTypeDutyCycle:                 ApiCommon.MetricType_DUTY_CYCLE,

	TypeGpuDutyCycle:                      ApiCommon.MetricType_DUTY_CYCLE,
	TypeGpuDutyCycleLowerBound:            ApiCommon.MetricType_DUTY_CYCLE,
	TypeGpuDutyCycleUpperBound:            ApiCommon.MetricType_DUTY_CYCLE,
	TypeGpuMemoryUsedBytes:                ApiCommon.MetricType_MEMORY_USAGE_BYTES,
	TypeGpuMemoryUsedBytesLowerBound:      ApiCommon.MetricType_MEMORY_USAGE_BYTES,
	TypeGpuMemoryUsedBytesUpperBound:      ApiCommon.MetricType_MEMORY_USAGE_BYTES,
	TypeGpuPowerUsageMilliWatts:           ApiCommon.MetricType_POWER_USAGE_WATTS,
	TypeGpuPowerUsageMilliWattsLowerBound: ApiCommon.MetricType_POWER_USAGE_WATTS,
	TypeGpuPowerUsageMilliWattsUpperBound: ApiCommon.MetricType_POWER_USAGE_WATTS,
	TypeGpuTemperatureCelsius:             ApiCommon.MetricType_TEMPERATURE_CELSIUS,
	TypeGpuTemperatureCelsiusLowerBound:   ApiCommon.MetricType_TEMPERATURE_CELSIUS,
	TypeGpuTemperatureCelsiusUpperBound:   ApiCommon.MetricType_TEMPERATURE_CELSIUS,
}
