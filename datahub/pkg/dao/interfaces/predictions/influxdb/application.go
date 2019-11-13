package influxdb

import (
	DaoPredictionTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
	RepoInfluxPrediction "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/predictions"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

type ApplicationPredictions struct {
	InfluxDBConfig InternalInflux.Config
}

func NewApplicationPredictionsWithConfig(config InternalInflux.Config) DaoPredictionTypes.ApplicationPredictionsDAO {
	return &ApplicationPredictions{InfluxDBConfig: config}
}

// CreateApplicationPredictions Implementation of prediction dao interface
func (p *ApplicationPredictions) CreatePredictions(predictions DaoPredictionTypes.ApplicationPredictionMap) error {
	predictionRepo := RepoInfluxPrediction.NewApplicationRepositoryWithConfig(p.InfluxDBConfig)

	err := predictionRepo.CreatePredictions(predictions)
	if err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}

func (p *ApplicationPredictions) ListPredictions(request DaoPredictionTypes.ListApplicationPredictionsRequest) (DaoPredictionTypes.ApplicationPredictionMap, error) {
	applicationPredictionMap := DaoPredictionTypes.NewApplicationPredictionMap()

	predictionRepo := RepoInfluxPrediction.NewApplicationRepositoryWithConfig(p.InfluxDBConfig)
	applicationPredictions, err := predictionRepo.ListPredictions(request)
	if err != nil {
		scope.Error(err.Error())
		return DaoPredictionTypes.NewApplicationPredictionMap(), err
	}
	for _, applicationPrediction := range applicationPredictions {
		applicationPredictionMap.AddApplicationPrediction(applicationPrediction)
	}

	return applicationPredictionMap, nil
}
