package influxdb

import (
	DaoPredictionTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
	RepoInfluxPrediction "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/predictions"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
)

type PodPredictions struct {
	InfluxDBConfig InternalInflux.Config
}

func NewPodPredictionsWithConfig(config InternalInflux.Config) DaoPredictionTypes.PodPredictionsDAO {
	return &PodPredictions{InfluxDBConfig: config}
}

func (p *PodPredictions) CreatePredictions(predictions DaoPredictionTypes.PodPredictionMap) error {
	containerRepo := RepoInfluxPrediction.NewContainerRepositoryWithConfig(p.InfluxDBConfig)

	predictionSampleList := make([]*DaoPredictionTypes.ContainerPredictionSample, 0)
	for _, podMetric := range predictions.MetricMap {
		podName := podMetric.ObjectMeta.Name
		namespace := podMetric.ObjectMeta.Namespace
		nodeName := podMetric.ObjectMeta.NodeName
		clusterName := podMetric.ObjectMeta.ClusterName

		for _, containerMetric := range podMetric.ContainerPredictionMap.MetricMap {
			containerName := containerMetric.ContainerName

			// Handle predicted raw data
			for metricType, metricData := range containerMetric.PredictionRaw {
				predictionSample := &DaoPredictionTypes.ContainerPredictionSample{
					ContainerName: containerName,
					PodName:       podName,
					Namespace:     namespace,
					NodeName:      nodeName,
					ClusterName:   clusterName,
					MetricType:    metricType,
					MetricKind:    FormatEnum.MetricKindRaw,
					Predictions:   metricData,
				}
				predictionSampleList = append(predictionSampleList, predictionSample)
			}

			// Handle predicted upper bound data
			for metricType, metricData := range containerMetric.PredictionUpperBound {
				predictionSample := &DaoPredictionTypes.ContainerPredictionSample{
					ContainerName: containerName,
					PodName:       podName,
					Namespace:     namespace,
					NodeName:      nodeName,
					ClusterName:   clusterName,
					MetricType:    metricType,
					MetricKind:    FormatEnum.MetricKindUpperBound,
					Predictions:   metricData,
				}
				predictionSampleList = append(predictionSampleList, predictionSample)
			}

			// Handle predicted lower bound data
			for metricType, metricData := range containerMetric.PredictionLowerBound {
				predictionSample := &DaoPredictionTypes.ContainerPredictionSample{
					ContainerName: containerName,
					PodName:       podName,
					Namespace:     namespace,
					NodeName:      nodeName,
					ClusterName:   clusterName,
					MetricType:    metricType,
					MetricKind:    FormatEnum.MetricKindLowerBound,
					Predictions:   metricData,
				}
				predictionSampleList = append(predictionSampleList, predictionSample)
			}
		}
	}

	err := containerRepo.CreatePredictions(predictionSampleList)
	if err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}

func (p *PodPredictions) ListPredictions(request DaoPredictionTypes.ListPodPredictionsRequest) (DaoPredictionTypes.PodPredictionMap, error) {
	podPredictionMap := DaoPredictionTypes.NewPodPredictionMap()

	predictionRepo := RepoInfluxPrediction.NewContainerRepositoryWithConfig(p.InfluxDBConfig)
	containerPredictions, err := predictionRepo.ListPredictions(request)
	if err != nil {
		scope.Error(err.Error())
		return DaoPredictionTypes.NewPodPredictionMap(), err
	}
	for _, containerPrediction := range containerPredictions {
		podPredictionMap.AddContainerPrediction(containerPrediction)
	}

	return podPredictionMap, nil
}

func (p *PodPredictions) FillPredictions(predictions []*ApiPredictions.PodPrediction, fillDays int64) error {
	// TODO: check if need to implement this function !!!
	/*for _, podPrediction := range predictions {
		for _, containerPrediction := range podPrediction.ContainerPredictions {
			for _, metricData := range containerPrediction.PredictedRawData {
				if len(metricData.Data) < 2 {
					continue
				}

				tempSampleList := make([]*ApiCommon.Sample, 0)
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
					tempSample := &ApiCommon.Sample{
						Time:     &timestamp.Timestamp{Seconds: a},
						NumValue: tempSampleList[tempIndex].NumValue,
					}
					metricData.Data = append(metricData.Data, tempSample)
					index++
				}
			}
		}
	}*/

	return nil
}
