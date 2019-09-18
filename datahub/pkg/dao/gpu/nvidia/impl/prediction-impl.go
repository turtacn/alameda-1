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
			}
		case DatahubMetric.TypeGpuMemoryUsedBytes:
			predictionRepo := RepoInfluxGpuPrediction.NewMemoryUsedBytesRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[DatahubMetric.TypeGpuMemoryUsedBytes])
			if err != nil {
				scope.Error(err.Error())
			}
		case DatahubMetric.TypeGpuPowerUsageMilliWatts:
			predictionRepo := RepoInfluxGpuPrediction.NewPowerUsageMilliWattsRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[DatahubMetric.TypeGpuPowerUsageMilliWatts])
			if err != nil {
				scope.Error(err.Error())
			}
		case DatahubMetric.TypeGpuTemperatureCelsius:
			predictionRepo := RepoInfluxGpuPrediction.NewTemperatureCelsiusRepositoryWithConfig(p.InfluxDBConfig)
			err = predictionRepo.CreatePredictions(predictions[DatahubMetric.TypeGpuTemperatureCelsius])
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
