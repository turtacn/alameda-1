package influxdb

import (
	DaoPredictionTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
	RepoInfluxPrediction "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/predictions"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

type ClusterPredictions struct {
	InfluxDBConfig InternalInflux.Config
}

func NewClusterPredictionsWithConfig(config InternalInflux.Config) DaoPredictionTypes.ClusterPredictionsDAO {
	return &ClusterPredictions{InfluxDBConfig: config}
}

// CreateClusterPredictions Implementation of prediction dao interface
func (p *ClusterPredictions) CreatePredictions(predictions DaoPredictionTypes.ClusterPredictionMap) error {
	predictionRepo := RepoInfluxPrediction.NewClusterRepositoryWithConfig(p.InfluxDBConfig)

	err := predictionRepo.CreatePredictions(predictions)
	if err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}

func (p *ClusterPredictions) ListPredictions(request DaoPredictionTypes.ListClusterPredictionsRequest) (DaoPredictionTypes.ClusterPredictionMap, error) {
	clusterPredictionMap := DaoPredictionTypes.NewClusterPredictionMap()

	predictionRepo := RepoInfluxPrediction.NewClusterRepositoryWithConfig(p.InfluxDBConfig)
	clusterPredictions, err := predictionRepo.ListPredictions(request)
	if err != nil {
		scope.Error(err.Error())
		return DaoPredictionTypes.NewClusterPredictionMap(), err
	}
	for _, clusterPrediction := range clusterPredictions {
		clusterPredictionMap.AddClusterPrediction(clusterPrediction)
	}

	return clusterPredictionMap, nil
}
