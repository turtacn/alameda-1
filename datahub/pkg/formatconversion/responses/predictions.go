package responses

import (
	DaoPredictionTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/predictions/types"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	ApiPredictions "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/predictions"
)

type NodePredictionExtended struct {
	*DaoPredictionTypes.NodePrediction
}

func (d *NodePredictionExtended) ProducePredictions() *ApiPredictions.NodePrediction {
	var (
		rawDataChan        = make(chan ApiPredictions.MetricData)
		upperBoundDataChan = make(chan ApiPredictions.MetricData)
		lowerBoundDataChan = make(chan ApiPredictions.MetricData)
		numOfGoroutine     = 0

		datahubNodePrediction ApiPredictions.NodePrediction
	)

	datahubNodePrediction = ApiPredictions.NodePrediction{
		ObjectMeta:  NewObjectMeta(&d.ObjectMeta),
		IsScheduled: d.IsScheduled,
	}

	// Handle prediction raw data
	numOfGoroutine = 0
	for metricType, samples := range d.PredictionRaw {
		if datahubMetricType, exist := FormatEnum.TypeToDatahubMetricType[metricType]; exist {
			numOfGoroutine++
			go producePredictionMetricDataFromSamples(datahubMetricType, samples.Granularity, samples.Data, rawDataChan)
		}
	}
	for i := 0; i < numOfGoroutine; i++ {
		receivedPredictionData := <-rawDataChan
		datahubNodePrediction.PredictedRawData = append(datahubNodePrediction.PredictedRawData, &receivedPredictionData)
	}

	// Handle prediction upper bound data
	numOfGoroutine = 0
	for metricType, samples := range d.PredictionUpperBound {
		if datahubMetricType, exist := FormatEnum.TypeToDatahubMetricType[metricType]; exist {
			numOfGoroutine++
			go producePredictionMetricDataFromSamples(datahubMetricType, samples.Granularity, samples.Data, upperBoundDataChan)
		}
	}
	for i := 0; i < numOfGoroutine; i++ {
		receivedPredictionData := <-upperBoundDataChan
		datahubNodePrediction.PredictedUpperboundData = append(datahubNodePrediction.PredictedUpperboundData, &receivedPredictionData)
	}

	// Handle prediction lower bound data
	numOfGoroutine = 0
	for metricType, samples := range d.PredictionLowerBound {
		if datahubMetricType, exist := FormatEnum.TypeToDatahubMetricType[metricType]; exist {
			numOfGoroutine++
			go producePredictionMetricDataFromSamples(datahubMetricType, samples.Granularity, samples.Data, lowerBoundDataChan)
		}
	}
	for i := 0; i < numOfGoroutine; i++ {
		receivedPredictionData := <-lowerBoundDataChan
		datahubNodePrediction.PredictedLowerboundData = append(datahubNodePrediction.PredictedLowerboundData, &receivedPredictionData)
	}

	return &datahubNodePrediction
}

type PodPredictionExtended struct {
	*DaoPredictionTypes.PodPrediction
}

func (p *PodPredictionExtended) ProducePredictions() *ApiPredictions.PodPrediction {
	datahubPodPrediction := ApiPredictions.PodPrediction{
		ObjectMeta: NewObjectMeta(&p.ObjectMeta),
	}

	for _, ptrContainerPrediction := range p.ContainerPredictionMap.MetricMap {
		containerPredictionExtended := ContainerPredictionExtended{ptrContainerPrediction}
		datahubContainerPrediction := containerPredictionExtended.ProducePredictions()
		datahubPodPrediction.ContainerPredictions = append(datahubPodPrediction.ContainerPredictions, datahubContainerPrediction)
	}

	return &datahubPodPrediction
}

type ContainerPredictionExtended struct {
	*DaoPredictionTypes.ContainerPrediction
}

func (c *ContainerPredictionExtended) ProducePredictions() *ApiPredictions.ContainerPrediction {
	var (
		rawDataChan        = make(chan ApiPredictions.MetricData)
		upperBoundDataChan = make(chan ApiPredictions.MetricData)
		lowerBoundDataChan = make(chan ApiPredictions.MetricData)
		numOfGoroutine     = 0

		datahubContainerPrediction ApiPredictions.ContainerPrediction
	)

	datahubContainerPrediction = ApiPredictions.ContainerPrediction{
		Name: string(c.ContainerName),
	}

	// Handle prediction raw data
	numOfGoroutine = 0
	for metricType, samples := range c.PredictionRaw {
		if datahubMetricType, exist := FormatEnum.TypeToDatahubMetricType[metricType]; exist {
			numOfGoroutine++
			go producePredictionMetricDataFromSamples(datahubMetricType, samples.Granularity, samples.Data, rawDataChan)
		}
	}
	for i := 0; i < numOfGoroutine; i++ {
		receivedPredictionData := <-rawDataChan
		datahubContainerPrediction.PredictedRawData = append(datahubContainerPrediction.PredictedRawData, &receivedPredictionData)
	}

	// Handle prediction upper bound data
	numOfGoroutine = 0
	for metricType, samples := range c.PredictionUpperBound {
		if datahubMetricType, exist := FormatEnum.TypeToDatahubMetricType[metricType]; exist {
			numOfGoroutine++
			go producePredictionMetricDataFromSamples(datahubMetricType, samples.Granularity, samples.Data, upperBoundDataChan)
		}
	}
	for i := 0; i < numOfGoroutine; i++ {
		receivedPredictionData := <-upperBoundDataChan
		datahubContainerPrediction.PredictedUpperboundData = append(datahubContainerPrediction.PredictedUpperboundData, &receivedPredictionData)
	}

	// Handle prediction lower bound data
	numOfGoroutine = 0
	for metricType, samples := range c.PredictionLowerBound {
		if datahubMetricType, exist := FormatEnum.TypeToDatahubMetricType[metricType]; exist {
			numOfGoroutine++
			go producePredictionMetricDataFromSamples(datahubMetricType, samples.Granularity, samples.Data, lowerBoundDataChan)
		}
	}
	for i := 0; i < numOfGoroutine; i++ {
		receivedPredictionData := <-lowerBoundDataChan
		datahubContainerPrediction.PredictedLowerboundData = append(datahubContainerPrediction.PredictedLowerboundData, &receivedPredictionData)
	}

	return &datahubContainerPrediction
}
