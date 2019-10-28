package nvidia

import (
	DaoGpu "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/gpu/influxdb"
	RepoInfluxGpuMetric "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/gpu/nvidia/metrics"
	RepoInfluxGpuPrediction "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/gpu/nvidia/predictions"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
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
		case FormatEnum.TypeGpuDutyCycle:
			predictionRepo := RepoInfluxGpuPrediction.NewDutyCycleRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[FormatEnum.TypeGpuDutyCycle])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case FormatEnum.TypeGpuDutyCycleLowerBound:
			predictionRepo := RepoInfluxGpuPrediction.NewDutyCycleLowerBoundRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[FormatEnum.TypeGpuDutyCycleLowerBound])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case FormatEnum.TypeGpuDutyCycleUpperBound:
			predictionRepo := RepoInfluxGpuPrediction.NewDutyCycleUpperBoundRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[FormatEnum.TypeGpuDutyCycleUpperBound])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case FormatEnum.TypeGpuMemoryUsedBytes:
			memoryUsageRepo := RepoInfluxGpuPrediction.NewMemoryUsagePercentageRepositoryWithConfig(p.InfluxDBConfig)
			err = memoryUsageRepo.CreatePredictions(p.buildMemoryUsagePrediction(predictions[FormatEnum.TypeGpuMemoryUsedBytes]))
			if err != nil {
				scope.Error(err.Error())
				break
			}

			memoryUsedRepo := RepoInfluxGpuPrediction.NewMemoryUsedBytesRepositoryWithConfig(p.InfluxDBConfig)
			err = memoryUsedRepo.CreatePredictions(predictions[FormatEnum.TypeGpuMemoryUsedBytes])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case FormatEnum.TypeGpuMemoryUsedBytesLowerBound:
			predictionRepo := RepoInfluxGpuPrediction.NewMemoryUsedBytesLowerBoundRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[FormatEnum.TypeGpuMemoryUsedBytesLowerBound])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case FormatEnum.TypeGpuMemoryUsedBytesUpperBound:
			predictionRepo := RepoInfluxGpuPrediction.NewMemoryUsedBytesUpperBoundRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[FormatEnum.TypeGpuMemoryUsedBytesUpperBound])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case FormatEnum.TypeGpuPowerUsageMilliWatts:
			predictionRepo := RepoInfluxGpuPrediction.NewPowerUsageMilliWattsRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[FormatEnum.TypeGpuPowerUsageMilliWatts])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case FormatEnum.TypeGpuPowerUsageMilliWattsLowerBound:
			predictionRepo := RepoInfluxGpuPrediction.NewPowerUsageMilliWattsLowerBoundRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[FormatEnum.TypeGpuPowerUsageMilliWattsLowerBound])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case FormatEnum.TypeGpuPowerUsageMilliWattsUpperBound:
			predictionRepo := RepoInfluxGpuPrediction.NewPowerUsageMilliWattsUpperBoundRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[FormatEnum.TypeGpuPowerUsageMilliWattsUpperBound])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case FormatEnum.TypeGpuTemperatureCelsius:
			predictionRepo := RepoInfluxGpuPrediction.NewTemperatureCelsiusRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[FormatEnum.TypeGpuTemperatureCelsius])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case FormatEnum.TypeGpuTemperatureCelsiusLowerBound:
			predictionRepo := RepoInfluxGpuPrediction.NewTemperatureCelsiusLowerBoundRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[FormatEnum.TypeGpuTemperatureCelsiusLowerBound])
			if err != nil {
				scope.Error(err.Error())
				break
			}
		case FormatEnum.TypeGpuTemperatureCelsiusUpperBound:
			predictionRepo := RepoInfluxGpuPrediction.NewTemperatureCelsiusUpperBoundRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[FormatEnum.TypeGpuTemperatureCelsiusUpperBound])
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

func (p Prediction) ListPredictions(host, minorNumber, modelId, predictionId, granularity string, condition *DBCommon.QueryCondition) (DaoGpu.GpuPredictionMap, error) {
	gpuPredictionMap := DaoGpu.NewGpuPredictionMap()

	granularityValue, _ := strconv.ParseInt(granularity, 10, 64)

	if DaoGpu.GpuMetricUsedMap[FormatEnum.TypeGpuDutyCycle] {
		// Pack duty cycle predictions
		dutyCycleRepo := RepoInfluxGpuPrediction.NewDutyCycleRepositoryWithConfig(p.InfluxDBConfig)
		dutyCyclePredictions, err := dutyCycleRepo.ListPredictions(host, minorNumber, modelId, predictionId, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range dutyCyclePredictions {
			sample := FormatTypes.PredictionSample{
				Timestamp:    predictions.Time,
				Value:        strconv.FormatFloat(*predictions.Value, 'f', -1, 64),
				ModelId:      *predictions.ModelId,
				PredictionId: *predictions.PredictionId,
			}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, FormatEnum.TypeGpuDutyCycle, sample)
		}

		// Pack duty cycle lower bound predictions
		dutyCycleLowerBoundRepo := RepoInfluxGpuPrediction.NewDutyCycleLowerBoundRepositoryWithConfig(p.InfluxDBConfig)
		dutyCyclePredictions, err = dutyCycleLowerBoundRepo.ListPredictions(host, minorNumber, modelId, predictionId, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range dutyCyclePredictions {
			sample := FormatTypes.PredictionSample{
				Timestamp:    predictions.Time,
				Value:        strconv.FormatFloat(*predictions.Value, 'f', -1, 64),
				ModelId:      *predictions.ModelId,
				PredictionId: *predictions.PredictionId,
			}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, FormatEnum.TypeGpuDutyCycleLowerBound, sample)
		}

		// Pack duty cycle upper bound predictions
		dutyCycleUpperBoundRepo := RepoInfluxGpuPrediction.NewDutyCycleUpperBoundRepositoryWithConfig(p.InfluxDBConfig)
		dutyCyclePredictions, err = dutyCycleUpperBoundRepo.ListPredictions(host, minorNumber, modelId, predictionId, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range dutyCyclePredictions {
			sample := FormatTypes.PredictionSample{
				Timestamp:    predictions.Time,
				Value:        strconv.FormatFloat(*predictions.Value, 'f', -1, 64),
				ModelId:      *predictions.ModelId,
				PredictionId: *predictions.PredictionId,
			}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, FormatEnum.TypeGpuDutyCycleUpperBound, sample)
		}
	}

	if DaoGpu.GpuMetricUsedMap[FormatEnum.TypeGpuMemoryUsedBytes] {
		// Pack memory used bytes predictions
		memoryUsedRepo := RepoInfluxGpuPrediction.NewMemoryUsedBytesRepositoryWithConfig(p.InfluxDBConfig)
		memoryUsedPredictions, err := memoryUsedRepo.ListPredictions(host, minorNumber, modelId, predictionId, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range memoryUsedPredictions {
			sample := FormatTypes.PredictionSample{
				Timestamp:    predictions.Time,
				Value:        strconv.FormatFloat(*predictions.Value, 'f', -1, 64),
				ModelId:      *predictions.ModelId,
				PredictionId: *predictions.PredictionId,
			}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, FormatEnum.TypeGpuMemoryUsedBytes, sample)
		}

		// Pack memory used bytes lower bound predictions
		memoryUsedLowerBoundRepo := RepoInfluxGpuPrediction.NewMemoryUsedBytesLowerBoundRepositoryWithConfig(p.InfluxDBConfig)
		memoryUsedPredictions, err = memoryUsedLowerBoundRepo.ListPredictions(host, minorNumber, modelId, predictionId, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range memoryUsedPredictions {
			sample := FormatTypes.PredictionSample{
				Timestamp:    predictions.Time,
				Value:        strconv.FormatFloat(*predictions.Value, 'f', -1, 64),
				ModelId:      *predictions.ModelId,
				PredictionId: *predictions.PredictionId,
			}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, FormatEnum.TypeGpuMemoryUsedBytesLowerBound, sample)
		}

		// Pack memory used bytes upper bound predictions
		memoryUsedUpperBoundRepo := RepoInfluxGpuPrediction.NewMemoryUsedBytesUpperBoundRepositoryWithConfig(p.InfluxDBConfig)
		memoryUsedPredictions, err = memoryUsedUpperBoundRepo.ListPredictions(host, minorNumber, modelId, predictionId, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range memoryUsedPredictions {
			sample := FormatTypes.PredictionSample{
				Timestamp:    predictions.Time,
				Value:        strconv.FormatFloat(*predictions.Value, 'f', -1, 64),
				ModelId:      *predictions.ModelId,
				PredictionId: *predictions.PredictionId,
			}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, FormatEnum.TypeGpuMemoryUsedBytesUpperBound, sample)
		}
	}

	if DaoGpu.GpuMetricUsedMap[FormatEnum.TypeGpuPowerUsageMilliWatts] {
		// Pack power usage milli watts predictions
		powerUsageRepo := RepoInfluxGpuPrediction.NewPowerUsageMilliWattsRepositoryWithConfig(p.InfluxDBConfig)
		poserUsagePredictions, err := powerUsageRepo.ListPredictions(host, minorNumber, modelId, predictionId, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range poserUsagePredictions {
			sample := FormatTypes.PredictionSample{
				Timestamp:    predictions.Time,
				Value:        strconv.FormatFloat(*predictions.Value, 'f', -1, 64),
				ModelId:      *predictions.ModelId,
				PredictionId: *predictions.PredictionId,
			}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, FormatEnum.TypeGpuPowerUsageMilliWatts, sample)
		}

		// Pack power usage milli watts lower bound predictions
		powerUsageLowerBoundRepo := RepoInfluxGpuPrediction.NewPowerUsageMilliWattsLowerBoundRepositoryWithConfig(p.InfluxDBConfig)
		poserUsagePredictions, err = powerUsageLowerBoundRepo.ListPredictions(host, minorNumber, modelId, predictionId, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range poserUsagePredictions {
			sample := FormatTypes.PredictionSample{
				Timestamp:    predictions.Time,
				Value:        strconv.FormatFloat(*predictions.Value, 'f', -1, 64),
				ModelId:      *predictions.ModelId,
				PredictionId: *predictions.PredictionId,
			}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, FormatEnum.TypeGpuPowerUsageMilliWattsLowerBound, sample)
		}

		// Pack power usage milli watts upper bound predictions
		powerUsageUpperBoundRepo := RepoInfluxGpuPrediction.NewPowerUsageMilliWattsUpperBoundRepositoryWithConfig(p.InfluxDBConfig)
		poserUsagePredictions, err = powerUsageUpperBoundRepo.ListPredictions(host, minorNumber, modelId, predictionId, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range poserUsagePredictions {
			sample := FormatTypes.PredictionSample{
				Timestamp:    predictions.Time,
				Value:        strconv.FormatFloat(*predictions.Value, 'f', -1, 64),
				ModelId:      *predictions.ModelId,
				PredictionId: *predictions.PredictionId,
			}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, FormatEnum.TypeGpuPowerUsageMilliWattsUpperBound, sample)
		}
	}

	if DaoGpu.GpuMetricUsedMap[FormatEnum.TypeGpuTemperatureCelsius] {
		// Pack temperature celsius predictions
		temperatureCelsiusRepo := RepoInfluxGpuPrediction.NewTemperatureCelsiusRepositoryWithConfig(p.InfluxDBConfig)
		temperatureCelsiusPredictions, err := temperatureCelsiusRepo.ListPredictions(host, minorNumber, modelId, predictionId, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range temperatureCelsiusPredictions {
			sample := FormatTypes.PredictionSample{
				Timestamp:    predictions.Time,
				Value:        strconv.FormatFloat(*predictions.Value, 'f', -1, 64),
				ModelId:      *predictions.ModelId,
				PredictionId: *predictions.PredictionId,
			}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, FormatEnum.TypeGpuTemperatureCelsius, sample)
		}

		// Pack temperature celsius lower bound predictions
		temperatureCelsiusLowerBoundRepo := RepoInfluxGpuPrediction.NewTemperatureCelsiusLowerBoundRepositoryWithConfig(p.InfluxDBConfig)
		temperatureCelsiusPredictions, err = temperatureCelsiusLowerBoundRepo.ListPredictions(host, minorNumber, modelId, predictionId, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range temperatureCelsiusPredictions {
			sample := FormatTypes.PredictionSample{
				Timestamp:    predictions.Time,
				Value:        strconv.FormatFloat(*predictions.Value, 'f', -1, 64),
				ModelId:      *predictions.ModelId,
				PredictionId: *predictions.PredictionId,
			}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, FormatEnum.TypeGpuTemperatureCelsiusLowerBound, sample)
		}

		// Pack temperature celsius upper bound predictions
		temperatureCelsiusUpperBoundRepo := RepoInfluxGpuPrediction.NewTemperatureCelsiusUpperBoundRepositoryWithConfig(p.InfluxDBConfig)
		temperatureCelsiusPredictions, err = temperatureCelsiusUpperBoundRepo.ListPredictions(host, minorNumber, modelId, predictionId, granularity, condition)
		if err != nil {
			return DaoGpu.NewGpuPredictionMap(), err
		}
		for _, predictions := range temperatureCelsiusPredictions {
			sample := FormatTypes.PredictionSample{
				Timestamp:    predictions.Time,
				Value:        strconv.FormatFloat(*predictions.Value, 'f', -1, 64),
				ModelId:      *predictions.ModelId,
				PredictionId: *predictions.PredictionId,
			}
			gpu := buildGpu(predictions.Name, predictions.Uuid, predictions.Host, predictions.Instance, predictions.Job, predictions.MinorNumber)
			gpuPredictionMap.AddGpuPrediction(gpu, granularityValue, FormatEnum.TypeGpuTemperatureCelsiusUpperBound, sample)
		}
	}

	return gpuPredictionMap, nil
}

func (p Prediction) buildMemoryUsagePrediction(memoryUsedBytesPredictions []*DaoGpu.GpuPrediction) []*DaoGpu.GpuPrediction {
	predictions := make([]*DaoGpu.GpuPrediction, 0)

	memoryTotalRepo := RepoInfluxGpuMetric.NewMemoryTotalBytesRepositoryWithConfig(p.InfluxDBConfig)
	memoryTotalMetrics, err := memoryTotalRepo.ListMemoryTotalBytes("", "", DBCommon.NewQueryCondition(0, 1, 0, 30))
	if err != nil {
		scope.Error(err.Error())
		scope.Error("failed to list memory total bytes when build memory usage percentage list")
		return make([]*DaoGpu.GpuPrediction, 0)
	}

	for _, memoryUsedBytes := range memoryUsedBytesPredictions {
		for _, memoryTotal := range memoryTotalMetrics {
			if memoryUsedBytes.Uuid == *memoryTotal.Uuid {
				gpuPrediction := DaoGpu.NewGpuPrediction()
				gpuPrediction.Name = memoryUsedBytes.Name
				gpuPrediction.Uuid = memoryUsedBytes.Uuid
				gpuPrediction.Metadata.Host = memoryUsedBytes.Metadata.Host
				gpuPrediction.Metadata.Instance = memoryUsedBytes.Metadata.Instance
				gpuPrediction.Metadata.Job = memoryUsedBytes.Metadata.Job
				gpuPrediction.Metadata.MinorNumber = memoryUsedBytes.Metadata.MinorNumber
				gpuPrediction.Granularity = memoryUsedBytes.Granularity

				for _, metric := range memoryUsedBytes.Metrics {
					usedBytes, _ := strconv.ParseFloat(metric.Value, 64)
					percentage := usedBytes / *memoryTotal.Value
					sample := FormatTypes.PredictionSample{Timestamp: metric.Timestamp, Value: strconv.FormatFloat(percentage, 'f', -1, 64)}

					gpuPrediction.Metrics = append(gpuPrediction.Metrics, sample)
				}
				predictions = append(predictions, gpuPrediction)

				break
			}
		}
	}

	return predictions
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
