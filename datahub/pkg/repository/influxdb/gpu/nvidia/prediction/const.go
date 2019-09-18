package prediction

import (
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

const (
	DutyCycle            InternalInflux.Measurement = "nvidia_gpu_duty_cycle"
	MemoryUsedBytes      InternalInflux.Measurement = "nvidia_gpu_memory_used_bytes"
	PowerUsageMilliWatts InternalInflux.Measurement = "nvidia_gpu_power_usage_milliwatts"
	TemperatureCelsius   InternalInflux.Measurement = "nvidia_gpu_temperature_celsius"
)
