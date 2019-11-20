package nvidia

import (
	DaoGpu "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/gpu/influxdb"
	RepoInfluxGpuMetric "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/gpu/nvidia/metrics"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
)

var (
	scope = Log.RegisterScope("dao_gpu_implement", "dao implement", 0)
)

type Gpu struct {
	InfluxDBConfig InternalInflux.Config
}

func NewGpuWithConfig(config InternalInflux.Config) DaoGpu.GpuDAO {
	return Gpu{InfluxDBConfig: config}
}

func (p Gpu) ListGpus(host, minorNumber string, condition *DBCommon.QueryCondition) ([]*DaoGpu.Gpu, error) {
	gpus := make([]*DaoGpu.Gpu, 0)

	queryCondition := &DBCommon.QueryCondition{
		StartTime:                 condition.StartTime,
		EndTime:                   condition.EndTime,
		Timeout:                   condition.Timeout,
		StepTime:                  condition.StepTime,
		TimestampOrder:            DBCommon.Desc,
		Limit:                     1,
		AggregateOverTimeFunction: DBCommon.None,
	}

	dutyCycleRepo := RepoInfluxGpuMetric.NewDutyCycleRepositoryWithConfig(p.InfluxDBConfig)
	dutyCycleMetrics, err := dutyCycleRepo.ListMetrics(host, minorNumber, queryCondition)
	if err != nil {
		return make([]*DaoGpu.Gpu, 0), err
	}
	for _, metrics := range dutyCycleMetrics {
		gpu := DaoGpu.NewGpu()
		gpu.Name = *metrics.Name
		gpu.Uuid = *metrics.Uuid
		gpu.Metadata.Host = *metrics.Host
		gpu.Metadata.Instance = *metrics.Instance
		gpu.Metadata.Job = *metrics.Job
		gpu.Metadata.MinorNumber = *metrics.MinorNumber

		gpus = append(gpus, gpu)
	}

	memoryTotalRepo := RepoInfluxGpuMetric.NewMemoryTotalBytesRepositoryWithConfig(p.InfluxDBConfig)
	memoryTotalMetrics, err := memoryTotalRepo.ListMemoryTotalBytes(host, minorNumber, queryCondition)
	if err != nil {
		return make([]*DaoGpu.Gpu, 0), err
	}
	for _, metrics := range memoryTotalMetrics {
		for _, gpu := range gpus {
			if *metrics.Uuid == gpu.Uuid {
				gpu.Spec.MemoryTotal = float32(*metrics.Value)
				break
			}
		}
	}

	return gpus, nil
}
