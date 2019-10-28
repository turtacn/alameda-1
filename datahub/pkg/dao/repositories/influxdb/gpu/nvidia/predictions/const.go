package predictions

import (
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

const (
	DutyCycle                      InternalInflux.Measurement = "nvidia_gpu_duty_cycle"
	DutyCycleLowerBound            InternalInflux.Measurement = "nvidia_gpu_duty_cycle_lower_bound"
	DutyCycleUpperBound            InternalInflux.Measurement = "nvidia_gpu_duty_cycle_upper_bound"
	MemoryUsagePercentage          InternalInflux.Measurement = "nvidia_gpu_memory_usage_percentage"
	MemoryUsedBytes                InternalInflux.Measurement = "nvidia_gpu_memory_used_bytes"
	MemoryUsedBytesLowerBound      InternalInflux.Measurement = "nvidia_gpu_memory_used_bytes_lower_bound"
	MemoryUsedBytesUpperBound      InternalInflux.Measurement = "nvidia_gpu_memory_used_bytes_upper_bound"
	PowerUsageMilliWatts           InternalInflux.Measurement = "nvidia_gpu_power_usage_milliwatts"
	PowerUsageMilliWattsLowerBound InternalInflux.Measurement = "nvidia_gpu_power_usage_milliwatts_lower_bound"
	PowerUsageMilliWattsUpperBound InternalInflux.Measurement = "nvidia_gpu_power_usage_milliwatts_upper_bound"
	TemperatureCelsius             InternalInflux.Measurement = "nvidia_gpu_temperature_celsius"
	TemperatureCelsiusLowerBound   InternalInflux.Measurement = "nvidia_gpu_temperature_celsius_lower_bound"
	TemperatureCelsiusUpperBound   InternalInflux.Measurement = "nvidia_gpu_temperature_celsius_upper_bound"
)
