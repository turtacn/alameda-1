package types

import (
	DaoPrediction "github.com/containers-ai/alameda/datahub/pkg/dao/prediction"
	Metric "github.com/containers-ai/alameda/datahub/pkg/metric"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

type PodPredictionExtended struct {
	*DaoPrediction.PodPrediction
}

func (p *PodPredictionExtended) ProducePredictions() *DatahubV1alpha1.PodPrediction {
	var (
		datahubPodPrediction DatahubV1alpha1.PodPrediction
	)

	datahubPodPrediction = DatahubV1alpha1.PodPrediction{
		NamespacedName: &DatahubV1alpha1.NamespacedName{
			Namespace: string(p.Namespace),
			Name:      string(p.PodName),
		},
	}

	for _, ptrContainerPrediction := range *p.ContainersPredictionMap {
		containerPredictionExtended := ContainerPredictionExtended{ptrContainerPrediction}
		datahubContainerPrediction := containerPredictionExtended.ProducePredictions()
		datahubPodPrediction.ContainerPredictions = append(datahubPodPrediction.ContainerPredictions, datahubContainerPrediction)
	}

	return &datahubPodPrediction
}

type ContainerPredictionExtended struct {
	*DaoPrediction.ContainerPrediction
}

func (c *ContainerPredictionExtended) ProducePredictions() *DatahubV1alpha1.ContainerPrediction {
	var (
		metricDataChan = make(chan DatahubV1alpha1.MetricData)
		numOfGoroutine = 0

		datahubContainerPrediction DatahubV1alpha1.ContainerPrediction
	)

	datahubContainerPrediction = DatahubV1alpha1.ContainerPrediction{
		Name: string(c.ContainerName),
	}

	for metricType, samples := range c.PredictionsRaw {
		if datahubMetricType, exist := Metric.TypeToDatahubMetricType[metricType]; exist {
			numOfGoroutine++
			go produceDatahubMetricDataFromSamples(datahubMetricType, samples, metricDataChan)
		}
	}

	for i := 0; i < numOfGoroutine; i++ {
		receivedPredictionData := <-metricDataChan
		datahubContainerPrediction.PredictedRawData = append(datahubContainerPrediction.PredictedRawData, &receivedPredictionData)
	}

	return &datahubContainerPrediction
}

type NodePredictionExtended struct {
	*DaoPrediction.NodePrediction
}

func (d *NodePredictionExtended) ProducePredictions() *DatahubV1alpha1.NodePrediction {
	var (
		metricDataChan = make(chan DatahubV1alpha1.MetricData)
		numOfGoroutine = 0

		datahubNodePrediction DatahubV1alpha1.NodePrediction
	)

	datahubNodePrediction = DatahubV1alpha1.NodePrediction{
		Name:        string(d.NodeName),
		IsScheduled: d.IsScheduled,
	}

	for metricType, samples := range d.Predictions {
		if datahubMetricType, exist := Metric.TypeToDatahubMetricType[metricType]; exist {
			numOfGoroutine++
			go produceDatahubMetricDataFromSamples(datahubMetricType, samples, metricDataChan)
		}
	}

	for i := 0; i < numOfGoroutine; i++ {
		receivedPredictionData := <-metricDataChan
		datahubNodePrediction.PredictedRawData = append(datahubNodePrediction.PredictedRawData, &receivedPredictionData)
	}

	return &datahubNodePrediction
}

type NodesPredictionMapExtended struct {
	*DaoPrediction.NodesPredictionMap
}

func (d *NodesPredictionMapExtended) ProducePredictions() []*DatahubV1alpha1.NodePrediction {
	var (
		datahubNodePredictions = make([]*DatahubV1alpha1.NodePrediction, 0)
	)

	for _, ptrIsScheduledNodePredictionMap := range *d.NodesPredictionMap {

		if ptrScheduledNodePrediction, exist := (*ptrIsScheduledNodePredictionMap)[true]; exist {

			scheduledNodePredictionExtended := NodePredictionExtended{ptrScheduledNodePrediction}
			sechduledDatahubNodePrediction := scheduledNodePredictionExtended.ProducePredictions()
			datahubNodePredictions = append(datahubNodePredictions, sechduledDatahubNodePrediction)
		}

		if noneScheduledNodePrediction, exist := (*ptrIsScheduledNodePredictionMap)[false]; exist {

			noneScheduledNodePredictionExtended := NodePredictionExtended{noneScheduledNodePrediction}
			noneSechduledDatahubNodePrediction := noneScheduledNodePredictionExtended.ProducePredictions()
			datahubNodePredictions = append(datahubNodePredictions, noneSechduledDatahubNodePrediction)
		}
	}

	return datahubNodePredictions
}
