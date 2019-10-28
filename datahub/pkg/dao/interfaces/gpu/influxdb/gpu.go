package influxdb

import (
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
)

var GpuMetricUsedMap = map[FormatEnum.GpuMetricType]bool{}

type Gpu struct {
	Name     string
	Uuid     string
	Metadata GpuMetadata
	Spec     GpuSpec
}

type GpuMetadata struct {
	Host        string
	Instance    string
	Job         string
	MinorNumber string
}

type GpuSpec struct {
	MemoryTotal float32
}

type GpuDAO interface {
	ListGpus(host, minorNumber string, condition *DBCommon.QueryCondition) ([]*Gpu, error)
}

func NewGpu() *Gpu {
	gpu := &Gpu{}
	return gpu
}

func init() {
	GpuMetricUsedMap[FormatEnum.TypeGpuDutyCycle] = true
	GpuMetricUsedMap[FormatEnum.TypeGpuMemoryUsedBytes] = true
	GpuMetricUsedMap[FormatEnum.TypeGpuPowerUsageMilliWatts] = false
	GpuMetricUsedMap[FormatEnum.TypeGpuTemperatureCelsius] = false
}
