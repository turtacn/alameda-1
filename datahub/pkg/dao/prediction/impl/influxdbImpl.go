package impl

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/prediction"
	RepoInfluxPrediction "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb/prediction"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
)

type influxDB struct {
	influxDBConfig InternalInflux.Config
}

// NewInfluxDBWithConfig Constructor of influxdb prediction dao
func NewInfluxDBWithConfig(config InternalInflux.Config) prediction.DAO {
	return influxDB{
		influxDBConfig: config,
	}
}

// CreateContainerPredictions Implementation of prediction dao interface
func (i influxDB) CreateContainerPredictions(in *datahub_v1alpha1.CreatePodPredictionsRequest) error {
	var (
		err            error
		predictionRepo *RepoInfluxPrediction.ContainerRepository
	)

	predictionRepo = RepoInfluxPrediction.NewContainerRepositoryWithConfig(i.influxDBConfig)

	err = predictionRepo.CreateContainerPrediction(in)
	if err != nil {
		return errors.Wrap(err, "create container prediction failed")
	}

	return nil
}

// ListPodPredictions Implementation of prediction dao interface
func (i influxDB) ListPodPredictions(request prediction.ListPodPredictionsRequest) ([]*datahub_v1alpha1.PodPrediction, error) {
	predictionRepo := RepoInfluxPrediction.NewContainerRepositoryWithConfig(i.influxDBConfig)
	return predictionRepo.ListContainerPredictionsByRequest(request)
}

func (i influxDB) FillPodPredictions(predictions []*datahub_v1alpha1.PodPrediction, fillDays int64) error {
	for _, podPrediction := range predictions {
		for _, containerPrediction := range podPrediction.ContainerPredictions {
			for _, metricData := range containerPrediction.PredictedRawData {
				if len(metricData.Data) < 2 {
					continue
				}

				tempSampleList := make([]*datahub_v1alpha1.Sample, 0)
				step := metricData.Data[1].Time.Seconds - metricData.Data[0].Time.Seconds

				if step <= 0 {
					continue
				}

				startTime := metricData.Data[len(metricData.Data)-1].Time.Seconds + step
				endTime := metricData.Data[0].Time.Seconds + 86400*fillDays
				for _, sample := range metricData.Data {
					tempSampleList = append(tempSampleList, sample)
				}

				index := 0
				for a := startTime; a <= endTime; a += step {
					tempIndex := index % len(tempSampleList)
					tempSample := &datahub_v1alpha1.Sample{
						Time:     &timestamp.Timestamp{Seconds: a},
						NumValue: tempSampleList[tempIndex].NumValue,
					}
					metricData.Data = append(metricData.Data, tempSample)
					index++
				}
			}
		}
	}

	return nil
}

// CreateNodePredictions Implementation of prediction dao interface
func (i influxDB) CreateNodePredictions(in *datahub_v1alpha1.CreateNodePredictionsRequest) error {
	predictionRepo := RepoInfluxPrediction.NewNodeRepositoryWithConfig(i.influxDBConfig)

	err := predictionRepo.CreateNodePrediction(in)
	if err != nil {
		return errors.Wrap(err, "create node prediction failed")
	}

	return nil
}

// ListNodePredictions Implementation of prediction dao interface
func (i influxDB) ListNodePredictions(request prediction.ListNodePredictionsRequest) ([]*datahub_v1alpha1.NodePrediction, error) {
	predictionRepo := RepoInfluxPrediction.NewNodeRepositoryWithConfig(i.influxDBConfig)
	return predictionRepo.ListNodePredictionsByRequest(request)
}
