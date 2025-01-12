package v1alpha1

import (
	DaoMetric "github.com/containers-ai/alameda/datahub/pkg/dao/metric"
	DaoPrediction "github.com/containers-ai/alameda/datahub/pkg/dao/prediction"
	Metric "github.com/containers-ai/alameda/datahub/pkg/metric"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
)

type daoPodMetricExtended struct {
	*DaoMetric.PodMetric
}

func (p daoPodMetricExtended) datahubPodMetric() *DatahubV1alpha1.PodMetric {

	var (
		datahubPodMetric DatahubV1alpha1.PodMetric
	)

	datahubPodMetric = DatahubV1alpha1.PodMetric{
		NamespacedName: &DatahubV1alpha1.NamespacedName{
			Namespace: string(p.Namespace),
			Name:      string(p.PodName),
		},
	}

	for _, containerMetric := range *p.ContainersMetricMap {
		containerMetricExtended := daoContainerMetricExtended{containerMetric}
		datahubContainerMetric := containerMetricExtended.datahubContainerMetric()
		datahubPodMetric.ContainerMetrics = append(datahubPodMetric.ContainerMetrics, datahubContainerMetric)
	}

	return &datahubPodMetric
}

type daoContainerMetricExtended struct {
	*DaoMetric.ContainerMetric
}

func (c daoContainerMetricExtended) datahubContainerMetric() *DatahubV1alpha1.ContainerMetric {

	var (
		metricDataChan  = make(chan DatahubV1alpha1.MetricData)
		numOfGoroutines = 0

		datahubContainerMetric DatahubV1alpha1.ContainerMetric
	)

	datahubContainerMetric = DatahubV1alpha1.ContainerMetric{
		Name: string(c.ContainerName),
	}

	for metricType, samples := range c.Metrics {
		if datahubMetricType, exist := Metric.TypeToDatahubMetricType[metricType]; exist {
			numOfGoroutines++
			go produceDatahubMetricDataFromSamples(datahubMetricType, samples, metricDataChan)
		}
	}

	for i := 0; i < numOfGoroutines; i++ {
		receivedMetricData := <-metricDataChan
		datahubContainerMetric.MetricData = append(datahubContainerMetric.MetricData, &receivedMetricData)
	}

	return &datahubContainerMetric
}

type daoNodeMetricExtended struct {
	*DaoMetric.NodeMetric
}

func (n daoNodeMetricExtended) datahubNodeMetric() *DatahubV1alpha1.NodeMetric {

	var (
		metricDataChan  = make(chan DatahubV1alpha1.MetricData)
		numOfGoroutines = 0

		datahubNodeMetric DatahubV1alpha1.NodeMetric
	)

	datahubNodeMetric = DatahubV1alpha1.NodeMetric{
		Name: n.NodeName,
	}

	for metricType, samples := range n.Metrics {
		if datahubMetricType, exist := Metric.TypeToDatahubMetricType[metricType]; exist {
			numOfGoroutines++
			go produceDatahubMetricDataFromSamples(datahubMetricType, samples, metricDataChan)
		}
	}

	for i := 0; i < numOfGoroutines; i++ {
		receivedMetricData := <-metricDataChan
		datahubNodeMetric.MetricData = append(datahubNodeMetric.MetricData, &receivedMetricData)
	}

	return &datahubNodeMetric
}

type daoPtrPodPredictionExtended struct {
	*DaoPrediction.PodPrediction
}

func (p daoPtrPodPredictionExtended) datahubPodPrediction() *DatahubV1alpha1.PodPrediction {

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
		containerPredictionExtended := daoContainerPredictionExtended{ptrContainerPrediction}
		datahubContainerPrediction := containerPredictionExtended.datahubContainerPrediction()
		datahubPodPrediction.ContainerPredictions = append(datahubPodPrediction.ContainerPredictions, datahubContainerPrediction)
	}

	return &datahubPodPrediction
}

type daoContainerPredictionExtended struct {
	*DaoPrediction.ContainerPrediction
}

func (c daoContainerPredictionExtended) datahubContainerPrediction() *DatahubV1alpha1.ContainerPrediction {

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

type daoPtrNodePredictionExtended struct {
	*DaoPrediction.NodePrediction
}

func (d daoPtrNodePredictionExtended) datahubNodePrediction() *DatahubV1alpha1.NodePrediction {

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

type daoPtrNodesPredictionMapExtended struct {
	*DaoPrediction.NodesPredictionMap
}

func (d daoPtrNodesPredictionMapExtended) datahubNodePredictions() []*DatahubV1alpha1.NodePrediction {

	var (
		datahubNodePredictions = make([]*DatahubV1alpha1.NodePrediction, 0)
	)

	for _, ptrIsScheduledNodePredictionMap := range *d.NodesPredictionMap {

		if ptrScheduledNodePrediction, exist := (*ptrIsScheduledNodePredictionMap)[true]; exist {

			scheduledNodePredictionExtended := daoPtrNodePredictionExtended{ptrScheduledNodePrediction}
			sechduledDatahubNodePrediction := scheduledNodePredictionExtended.datahubNodePrediction()
			datahubNodePredictions = append(datahubNodePredictions, sechduledDatahubNodePrediction)
		}

		if noneScheduledNodePrediction, exist := (*ptrIsScheduledNodePredictionMap)[false]; exist {

			noneScheduledNodePredictionExtended := daoPtrNodePredictionExtended{noneScheduledNodePrediction}
			noneSechduledDatahubNodePrediction := noneScheduledNodePredictionExtended.datahubNodePrediction()
			datahubNodePredictions = append(datahubNodePredictions, noneSechduledDatahubNodePrediction)
		}
	}

	return datahubNodePredictions
}

func produceDatahubMetricDataFromSamples(metricType DatahubV1alpha1.MetricType, samples []Metric.Sample, MetricDataChan chan<- DatahubV1alpha1.MetricData) {

	var (
		datahubMetricData DatahubV1alpha1.MetricData
	)

	datahubMetricData = DatahubV1alpha1.MetricData{
		MetricType: metricType,
	}

	for _, sample := range samples {

		// TODO: Send error to caller
		googleTimestamp, err := ptypes.TimestampProto(sample.Timestamp)
		if err != nil {
			googleTimestamp = nil
		}

		datahubSample := DatahubV1alpha1.Sample{Time: googleTimestamp, NumValue: sample.Value}
		datahubMetricData.Data = append(datahubMetricData.Data, &datahubSample)
	}

	MetricDataChan <- datahubMetricData
}
