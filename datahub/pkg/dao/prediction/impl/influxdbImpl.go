package impl

import (
	"errors"
	"github.com/containers-ai/alameda/datahub/pkg/dao/prediction"
	influxdb_container_preditcion_entity "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/prediction/container"
	influxdb_repository "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	influxdb_repository_preditcion "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb/prediction"
)

type influxDB struct {
	influxDBConfig influxdb_repository.Config
}

// NewInfluxDBWithConfig Constructor of influxdb prediction dao
func NewInfluxDBWithConfig(config influxdb_repository.Config) prediction.DAO {
	return influxDB{
		influxDBConfig: config,
	}
}

// CreateContainerPredictions Implementation of prediction dao interface
func (i influxDB) CreateContainerPredictions(containerPredictions []*prediction.ContainerPrediction) error {

	var (
		err error

		predictionRepo *influxdb_repository_preditcion.ContainerRepository
	)

	predictionRepo = influxdb_repository_preditcion.NewContainerRepositoryWithConfig(i.influxDBConfig)

	err = predictionRepo.CreateContainerPrediction(containerPredictions)
	if err != nil {
		return errors.New("create container prediction failed: ")
	}

	return nil
}

// ListPodPredictions Implementation of prediction dao interface
func (i influxDB) ListPodPredictions(request prediction.ListPodPredictionsRequest) (prediction.PodsPredictionMap, error) {

	var (
		err error

		predictionRepo                      *influxdb_repository_preditcion.ContainerRepository
		influxDBContainerPredictionEntities []*influxdb_container_preditcion_entity.Entity
		podsPredictionMap                   *prediction.PodsPredictionMap
	)

	podsPredictionMap = &prediction.PodsPredictionMap{}
	predictionRepo = influxdb_repository_preditcion.NewContainerRepositoryWithConfig(i.influxDBConfig)

	influxDBContainerPredictionEntities, err = predictionRepo.ListContainerPredictionsByRequest(request)
	if err != nil {
		return *podsPredictionMap, errors.New("create container prediction failed: ")
	}

	for _, entity := range influxDBContainerPredictionEntities {

		containerPrediction := entity.ContainerPrediction()
		podsPredictionMap.AddContainerPrediction(containerPrediction)
	}

	return *podsPredictionMap, nil
}
