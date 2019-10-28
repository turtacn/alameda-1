package influxdb

import (
	DaoPredictionTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
	RepoInfluxPrediction "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/predictions"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
)

var (
	scope = Log.RegisterScope("dao_influxdb_prediction_implement", "dao implement", 0)
)

type NodePredictions struct {
	InfluxDBConfig InternalInflux.Config
}

func NewNodePredictionsWithConfig(config InternalInflux.Config) DaoPredictionTypes.NodePredictionsDAO {
	return &NodePredictions{InfluxDBConfig: config}
}

// CreateNodePredictions Implementation of prediction dao interface
func (p *NodePredictions) CreatePredictions(predictions DaoPredictionTypes.NodePredictionMap) error {
	predictionRepo := RepoInfluxPrediction.NewNodeRepositoryWithConfig(p.InfluxDBConfig)

	err := predictionRepo.CreatePredictions(predictions)
	if err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}

func (p *NodePredictions) ListPredictions(request DaoPredictionTypes.ListNodePredictionsRequest) (DaoPredictionTypes.NodePredictionMap, error) {
	nodePredictionMap := DaoPredictionTypes.NewNodePredictionMap()

	predictionRepo := RepoInfluxPrediction.NewNodeRepositoryWithConfig(p.InfluxDBConfig)
	nodePredictions, err := predictionRepo.ListPredictions(request)
	if err != nil {
		scope.Error(err.Error())
		return DaoPredictionTypes.NewNodePredictionMap(), err
	}
	for _, nodePrediction := range nodePredictions {
		nodePredictionMap.AddNodePrediction(nodePrediction)
	}

	return nodePredictionMap, nil
}

func (p *NodePredictions) FillPredictions(predictions []*ApiPredictions.NodePrediction, fillDays int64) error {
	return nil
}
