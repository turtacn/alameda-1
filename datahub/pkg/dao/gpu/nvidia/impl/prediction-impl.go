package impl

import (
	DaoGpu "github.com/containers-ai/alameda/datahub/pkg/dao/gpu/nvidia"
	DatahubMetric "github.com/containers-ai/alameda/datahub/pkg/metric"
	RepoInfluxGpuPrediction "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb/gpu/nvidia/prediction"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	"strconv"
)

type Prediction struct {
	InfluxDBConfig InternalInflux.Config
}

func NewPredictionWithConfig(config InternalInflux.Config) DaoGpu.PredictionsDAO {
	return Prediction{InfluxDBConfig: config}
}

func (p Prediction) CreatePredictions(predictions DaoGpu.GpuPredictionMap) error {
	for k := range predictions {
		var err error

		switch k {
		case DatahubMetric.TypeGpuDutyCycle:
			predictionRepo := RepoInfluxGpuPrediction.NewDutyCycleRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[DatahubMetric.TypeGpuDutyCycle])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case DatahubMetric.TypeGpuDutyCycleLowerBound:
			predictionRepo := RepoInfluxGpuPrediction.NewDutyCycleLowerBoundRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[DatahubMetric.TypeGpuDutyCycleLowerBound])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case DatahubMetric.TypeGpuDutyCycleUpperBound:
			predictionRepo := RepoInfluxGpuPrediction.NewDutyCycleUpperBoundRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[DatahubMetric.TypeGpuDutyCycleUpperBound])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case DatahubMetric.TypeGpuMemoryUsedBytes:
			predictionRepo := RepoInfluxGpuPrediction.NewMemoryUsedBytesRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[DatahubMetric.TypeGpuMemoryUsedBytes])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case DatahubMetric.TypeGpuMemoryUsedBytesLowerBound:
			predictionRepo := RepoInfluxGpuPrediction.NewMemoryUsedBytesLowerBoundRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[DatahubMetric.TypeGpuMemoryUsedBytesLowerBound])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case DatahubMetric.TypeGpuMemoryUsedBytesUpperBound:
			predictionRepo := RepoInfluxGpuPrediction.NewMemoryUsedBytesUpperBoundRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[DatahubMetric.TypeGpuMemoryUsedBytesUpperBound])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case DatahubMetric.TypeGpuPowerUsageMilliWatts:
			predictionRepo := RepoInfluxGpuPrediction.NewPowerUsageMilliWattsRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[DatahubMetric.TypeGpuPowerUsageMilliWatts])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case DatahubMetric.TypeGpuPowerUsageMilliWattsLowerBound:
			predictionRepo := RepoInfluxGpuPrediction.NewPowerUsageMilliWattsLowerBoundRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[DatahubMetric.TypeGpuPowerUsageMilliWattsLowerBound])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case DatahubMetric.TypeGpuPowerUsageMilliWattsUpperBound:
			predictionRepo := RepoInfluxGpuPrediction.NewPowerUsageMilliWattsUpperBoundRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[DatahubMetric.TypeGpuPowerUsageMilliWattsUpperBound])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case DatahubMetric.TypeGpuTemperatureCelsius:
			predictionRepo := RepoInfluxGpuPrediction.NewTemperatureCelsiusRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[DatahubMetric.TypeGpuTemperatureCelsius])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case DatahubMetric.TypeGpuTemperatureCelsiusLowerBound:
			predictionRepo := RepoInfluxGpuPrediction.NewTemperatureCelsiusLowerBoundRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[DatahubMetric.TypeGpuTemperatureCelsiusLowerBound])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case DatahubMetric.TypeGpuTemperatureCelsiusUpperBound:
			predictionRepo := RepoInfluxGpuPrediction.NewTemperatureCelsiusUpperBoundRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[DatahubMetric.TypeGpuTemperatureCelsiusUpperBound])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		}

		if err != nil {
			scope.Error("failed to create gpu predictions")
			return err
		}
	}

	return nil
}

func (p Prediction) ListPredictions(host, minorNumber, granularity string, condition *DBCommon.QueryCondition) (DaoGpu.GpuPredictionMap, error) {
	gpuPredictionMap := DaoGpu.NewGpuPredictionMap()

	granularityValue, _ := strconv.ParseInt(granularity, 10, 64)

	if DaoGpu.GpuMetricUsedMap[DatahubMetric.TypeGpuDutyCycle] {
		// Pack duty cycle predictions
		dutyCycleRepo := RepoInfluxGpuPrediction.NewDutyCycleRepositoryWithConfig(p.InfluxDBConfig)
		dutyCyclePredictions, err := dutyCycleRepo.ListPredictions(host, minorNumber, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range dutyCyclePredictions {
			sample := DatahubMetric.Sample{Timestamp: predictions.Time, Value: strconv.FormatFloat(*predictions.Value, 'f', -1, 64)}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, DatahubMetric.TypeGpuDutyCycle, sample)
		}

		// Pack duty cycle lower bound predictions
		dutyCycleLowerBoundRepo := RepoInfluxGpuPrediction.NewDutyCycleLowerBoundRepositoryWithConfig(p.InfluxDBConfig)
		dutyCyclePredictions, err = dutyCycleLowerBoundRepo.ListPredictions(host, minorNumber, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range dutyCyclePredictions {
			sample := DatahubMetric.Sample{Timestamp: predictions.Time, Value: strconv.FormatFloat(*predictions.Value, 'f', -1, 64)}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, DatahubMetric.TypeGpuDutyCycleLowerBound, sample)
		}

		// Pack duty cycle upper bound predictions
		dutyCycleUpperBoundRepo := RepoInfluxGpuPrediction.NewDutyCycleUpperBoundRepositoryWithConfig(p.InfluxDBConfig)
		dutyCyclePredictions, err = dutyCycleUpperBoundRepo.ListPredictions(host, minorNumber, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range dutyCyclePredictions {
			sample := DatahubMetric.Sample{Timestamp: predictions.Time, Value: strconv.FormatFloat(*predictions.Value, 'f', -1, 64)}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, DatahubMetric.TypeGpuDutyCycleUpperBound, sample)
		}
	}

	if DaoGpu.GpuMetricUsedMap[DatahubMetric.TypeGpuMemoryUsedBytes] {
		// Pack memory used bytes predictions
		memoryUsedRepo := RepoInfluxGpuPrediction.NewMemoryUsedBytesRepositoryWithConfig(p.InfluxDBConfig)
		memoryUsedPredictions, err := memoryUsedRepo.ListPredictions(host, minorNumber, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range memoryUsedPredictions {
			sample := DatahubMetric.Sample{Timestamp: predictions.Time, Value: strconv.FormatFloat(*predictions.Value, 'f', -1, 64)}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, DatahubMetric.TypeGpuMemoryUsedBytes, sample)
		}

		// Pack memory used bytes lower bound predictions
		memoryUsedLowerBoundRepo := RepoInfluxGpuPrediction.NewMemoryUsedBytesLowerBoundRepositoryWithConfig(p.InfluxDBConfig)
		memoryUsedPredictions, err = memoryUsedLowerBoundRepo.ListPredictions(host, minorNumber, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range memoryUsedPredictions {
			sample := DatahubMetric.Sample{Timestamp: predictions.Time, Value: strconv.FormatFloat(*predictions.Value, 'f', -1, 64)}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, DatahubMetric.TypeGpuMemoryUsedBytesLowerBound, sample)
		}

		// Pack memory used bytes upper bound predictions
		memoryUsedUpperBoundRepo := RepoInfluxGpuPrediction.NewMemoryUsedBytesUpperBoundRepositoryWithConfig(p.InfluxDBConfig)
		memoryUsedPredictions, err = memoryUsedUpperBoundRepo.ListPredictions(host, minorNumber, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range memoryUsedPredictions {
			sample := DatahubMetric.Sample{Timestamp: predictions.Time, Value: strconv.FormatFloat(*predictions.Value, 'f', -1, 64)}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, DatahubMetric.TypeGpuMemoryUsedBytesUpperBound, sample)
		}
	}

	if DaoGpu.GpuMetricUsedMap[DatahubMetric.TypeGpuPowerUsageMilliWatts] {
		// Pack power usage milli watts predictions
		powerUsageRepo := RepoInfluxGpuPrediction.NewPowerUsageMilliWattsRepositoryWithConfig(p.InfluxDBConfig)
		poserUsagePredictions, err := powerUsageRepo.ListPredictions(host, minorNumber, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range poserUsagePredictions {
			sample := DatahubMetric.Sample{Timestamp: predictions.Time, Value: strconv.FormatFloat(*predictions.Value, 'f', -1, 64)}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, DatahubMetric.TypeGpuPowerUsageMilliWatts, sample)
		}

		// Pack power usage milli watts lower bound predictions
		powerUsageLowerBoundRepo := RepoInfluxGpuPrediction.NewPowerUsageMilliWattsLowerBoundRepositoryWithConfig(p.InfluxDBConfig)
		poserUsagePredictions, err = powerUsageLowerBoundRepo.ListPredictions(host, minorNumber, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range poserUsagePredictions {
			sample := DatahubMetric.Sample{Timestamp: predictions.Time, Value: strconv.FormatFloat(*predictions.Value, 'f', -1, 64)}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, DatahubMetric.TypeGpuPowerUsageMilliWattsLowerBound, sample)
		}

		// Pack power usage milli watts upper bound predictions
		powerUsageUpperBoundRepo := RepoInfluxGpuPrediction.NewPowerUsageMilliWattsUpperBoundRepositoryWithConfig(p.InfluxDBConfig)
		poserUsagePredictions, err = powerUsageUpperBoundRepo.ListPredictions(host, minorNumber, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range poserUsagePredictions {
			sample := DatahubMetric.Sample{Timestamp: predictions.Time, Value: strconv.FormatFloat(*predictions.Value, 'f', -1, 64)}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, DatahubMetric.TypeGpuPowerUsageMilliWattsUpperBound, sample)
		}
	}

	if DaoGpu.GpuMetricUsedMap[DatahubMetric.TypeGpuTemperatureCelsius] {
		// Pack temperature celsius predictions
		temperatureCelsiusRepo := RepoInfluxGpuPrediction.NewTemperatureCelsiusRepositoryWithConfig(p.InfluxDBConfig)
		temperatureCelsiusPredictions, err := temperatureCelsiusRepo.ListPredictions(host, minorNumber, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range temperatureCelsiusPredictions {
			sample := DatahubMetric.Sample{Timestamp: predictions.Time, Value: strconv.FormatFloat(*predictions.Value, 'f', -1, 64)}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, DatahubMetric.TypeGpuTemperatureCelsius, sample)
		}

		// Pack temperature celsius lower bound predictions
		temperatureCelsiusLowerBoundRepo := RepoInfluxGpuPrediction.NewTemperatureCelsiusLowerBoundRepositoryWithConfig(p.InfluxDBConfig)
		temperatureCelsiusPredictions, err = temperatureCelsiusLowerBoundRepo.ListPredictions(host, minorNumber, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range temperatureCelsiusPredictions {
			sample := DatahubMetric.Sample{Timestamp: predictions.Time, Value: strconv.FormatFloat(*predictions.Value, 'f', -1, 64)}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, DatahubMetric.TypeGpuTemperatureCelsiusLowerBound, sample)
		}

		// Pack temperature celsius upper bound predictions
		temperatureCelsiusUpperBoundRepo := RepoInfluxGpuPrediction.NewTemperatureCelsiusUpperBoundRepositoryWithConfig(p.InfluxDBConfig)
		temperatureCelsiusPredictions, err = temperatureCelsiusUpperBoundRepo.ListPredictions(host, minorNumber, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range temperatureCelsiusPredictions {
			sample := DatahubMetric.Sample{Timestamp: predictions.Time, Value: strconv.FormatFloat(*predictions.Value, 'f', -1, 64)}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, DatahubMetric.TypeGpuTemperatureCelsiusUpperBound, sample)
		}
	}

	return gpuPredictionMap, nil
}

func buildGpu(name, uuid, host, instance, job, minorNumber *string) *DaoGpu.Gpu {
	gpu := DaoGpu.NewGpu()

	if name != nil {
		gpu.Name = *name
	}
	if uuid != nil {
		gpu.Uuid = *uuid
	}
	if host != nil {
		gpu.Metadata.Host = *host
	}
	if instance != nil {
		gpu.Metadata.Instance = *instance
	}
	if job != nil {
		gpu.Metadata.Job = *job
	}
	if minorNumber != nil {
		gpu.Metadata.MinorNumber = *minorNumber
	}

	return gpu
}
