package datahub

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/metric"
	"github.com/containers-ai/alameda/datahub/pkg/dao/prediction"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
)

type daoPodMetricExtended metric.PodMetric

func (p daoPodMetricExtended) datahubPodMetric() datahub_v1alpha1.PodMetric {

	var (
		datahubPodMetric datahub_v1alpha1.PodMetric
	)

	datahubPodMetric = datahub_v1alpha1.PodMetric{
		NamespacedName: &datahub_v1alpha1.NamespacedName{
			Namespace: string(p.Namespace),
			Name:      string(p.PodName),
		},
	}

	for _, containerMetric := range p.ContainersMetricMap {
		containerMetricExtended := daoContainerMetricExtended(containerMetric)
		datahubContainerMetric := containerMetricExtended.datahubContainerMetric()
		datahubPodMetric.ContainerMetrics = append(datahubPodMetric.ContainerMetrics, &datahubContainerMetric)
	}

	return datahubPodMetric
}

type daoContainerMetricExtended metric.ContainerMetric

func (c daoContainerMetricExtended) NumberOfDatahubMetricDataNeededProducing() int {
	return 2
}

func (c daoContainerMetricExtended) datahubContainerMetric() datahub_v1alpha1.ContainerMetric {

	var (
		metricDataChan = make(chan datahub_v1alpha1.MetricData)

		datahubContainerMetric datahub_v1alpha1.ContainerMetric
	)

	datahubContainerMetric = datahub_v1alpha1.ContainerMetric{
		Name: string(c.ContainerName),
	}

	go produceDatahubMetricDataFromSamples(datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE, c.CPUMetircs, metricDataChan)
	go produceDatahubMetricDataFromSamples(datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES, c.MemoryMetrics, metricDataChan)

	for i := 0; i < c.NumberOfDatahubMetricDataNeededProducing(); i++ {
		receivedMetricData := <-metricDataChan
		datahubContainerMetric.MetricData = append(datahubContainerMetric.MetricData, &receivedMetricData)
	}

	return datahubContainerMetric
}

type daoNodeMetricExtended metric.NodeMetric

func (n daoNodeMetricExtended) NumberOfDatahubMetricDataNeededProducing() int {
	return 2
}

func (n daoNodeMetricExtended) datahubNodeMetric() datahub_v1alpha1.NodeMetric {

	var (
		metricDataChan = make(chan datahub_v1alpha1.MetricData)

		datahubNodeMetric datahub_v1alpha1.NodeMetric
	)

	datahubNodeMetric = datahub_v1alpha1.NodeMetric{
		Name: n.NodeName,
	}

	go produceDatahubMetricDataFromSamples(datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE, n.CPUUsageMetircs, metricDataChan)
	go produceDatahubMetricDataFromSamples(datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES, n.MemoryUsageMetrics, metricDataChan)

	for i := 0; i < n.NumberOfDatahubMetricDataNeededProducing(); i++ {
		receivedMetricData := <-metricDataChan
		datahubNodeMetric.MetricData = append(datahubNodeMetric.MetricData, &receivedMetricData)
	}

	return datahubNodeMetric
}

type daoPodPredictionExtended prediction.PodPrediction

func (p daoPodPredictionExtended) datahubPodPrediction() datahub_v1alpha1.PodPrediction {

	var (
		datahubPodPrediction datahub_v1alpha1.PodPrediction
	)

	datahubPodPrediction = datahub_v1alpha1.PodPrediction{
		NamespacedName: &datahub_v1alpha1.NamespacedName{
			Namespace: string(p.Namespace),
			Name:      string(p.PodName),
		},
	}

	for _, containerPrediction := range p.ContainersPredictionMap {
		containerPredictionExtended := daoContainerPredictionExtended(containerPrediction)
		datahubContainerPrediction := containerPredictionExtended.datahubContainerPrediction()
		datahubPodPrediction.ContainerPredictions = append(datahubPodPrediction.ContainerPredictions, &datahubContainerPrediction)
	}

	return datahubPodPrediction
}

type daoContainerPredictionExtended prediction.ContainerPrediction

func (c daoContainerPredictionExtended) NumberOfDatahubPredictionDataNeededProducing() int {
	return 2
}

func (c daoContainerPredictionExtended) datahubContainerPrediction() datahub_v1alpha1.ContainerPrediction {

	var (
		MetricDataChan = make(chan datahub_v1alpha1.MetricData)

		datahubContainerPrediction datahub_v1alpha1.ContainerPrediction
	)

	datahubContainerPrediction = datahub_v1alpha1.ContainerPrediction{
		Name: string(c.ContainerName),
	}

	go produceDatahubMetricDataFromSamples(datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE, c.CPUPredictions, MetricDataChan)
	go produceDatahubMetricDataFromSamples(datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES, c.MemoryPredictions, MetricDataChan)

	for i := 0; i < c.NumberOfDatahubPredictionDataNeededProducing(); i++ {
		receivedPredictionData := <-MetricDataChan
		datahubContainerPrediction.PredictedRawData = append(datahubContainerPrediction.PredictedRawData, &receivedPredictionData)
	}

	return datahubContainerPrediction
}

type daoNodePredictionExtended prediction.NodePrediction

func (d daoNodePredictionExtended) NumberOfDatahubPredictionDataNeededProducing() int {
	return 2
}

func (d daoNodePredictionExtended) datahubNodePrediction() datahub_v1alpha1.NodePrediction {

	var (
		MetricDataChan = make(chan datahub_v1alpha1.MetricData)

		datahubNodePrediction datahub_v1alpha1.NodePrediction
	)

	datahubNodePrediction = datahub_v1alpha1.NodePrediction{
		Name:        string(d.NodeName),
		IsScheduled: d.IsScheduled,
	}

	go produceDatahubMetricDataFromSamples(datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE, d.CPUUsagePredictions, MetricDataChan)
	go produceDatahubMetricDataFromSamples(datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES, d.MemoryUsagePredictions, MetricDataChan)

	for i := 0; i < d.NumberOfDatahubPredictionDataNeededProducing(); i++ {
		receivedPredictionData := <-MetricDataChan
		datahubNodePrediction.PredictedRawData = append(datahubNodePrediction.PredictedRawData, &receivedPredictionData)
	}

	return datahubNodePrediction
}

type daoNodesPredictionMapExtended prediction.NodesPredictionMap

func (d daoNodesPredictionMapExtended) datahubNodePredictions() []*datahub_v1alpha1.NodePrediction {

	var (
		datahubNodePredictions = make([]*datahub_v1alpha1.NodePrediction, 0)
	)

	for _, isScheduledNodePredictionMap := range d {

		if scheduledNodePrediction, exist := isScheduledNodePredictionMap[true]; exist {

			scheduledNodePredictionExtended := daoNodePredictionExtended(scheduledNodePrediction)
			sechduledDatahubNodePrediction := scheduledNodePredictionExtended.datahubNodePrediction()
			datahubNodePredictions = append(datahubNodePredictions, &sechduledDatahubNodePrediction)
		}

		if noneScheduledNodePrediction, exist := isScheduledNodePredictionMap[false]; exist {

			noneScheduledNodePredictionExtended := daoNodePredictionExtended(noneScheduledNodePrediction)
			noneSechduledDatahubNodePrediction := noneScheduledNodePredictionExtended.datahubNodePrediction()
			datahubNodePredictions = append(datahubNodePredictions, &noneSechduledDatahubNodePrediction)
		}
	}

	return datahubNodePredictions
}

func produceDatahubMetricDataFromSamples(metricType datahub_v1alpha1.MetricType, samples []metric.Sample, MetricDataChan chan<- datahub_v1alpha1.MetricData) {

	var (
		datahubMetricData datahub_v1alpha1.MetricData
	)

	datahubMetricData = datahub_v1alpha1.MetricData{
		MetricType: metricType,
	}

	for _, sample := range samples {

		// TODO: Send error to caller
		googleTimestamp, err := ptypes.TimestampProto(sample.Timestamp)
		if err != nil {
			googleTimestamp = nil
		}

		datahubSample := datahub_v1alpha1.Sample{Time: googleTimestamp, NumValue: sample.Value}
		datahubMetricData.Data = append(datahubMetricData.Data, &datahubSample)
	}

	MetricDataChan <- datahubMetricData
}
