package impl

import (
	DaoGpu "github.com/containers-ai/alameda/datahub/pkg/dao/gpu/nvidia"
	DatahubMetric "github.com/containers-ai/alameda/datahub/pkg/metric"
	RepoInfluxGpuPrediction "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb/gpu/nvidia/prediction"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
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
