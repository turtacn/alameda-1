package datahub

import (
	"errors"

	"github.com/containers-ai/alameda/datahub/pkg/dao/metric"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/timestamp"
)

type listPodMetricsRequestExtended datahub_v1alpha1.ListPodMetricsRequest

func (r listPodMetricsRequestExtended) validate() error {

	var (
		startTime *timestamp.Timestamp
		endTime   *timestamp.Timestamp
	)

	if r.TimeRange == nil {
		return errors.New("field \"time_range\" cannot be empty")
	}

	startTime = r.TimeRange.StartTime
	endTime = r.TimeRange.EndTime
	if startTime == nil || endTime == nil {
		return errors.New("field \"start_time\" and \"end_time\"  cannot be empty")
	}

	if startTime.Seconds+int64(startTime.Nanos) >= endTime.Seconds+int64(endTime.Nanos) {
		return errors.New("\"end_time\" must not be before \"start_time\"")
	}

	return nil
}

type listNodeMetricsRequestExtended datahub_v1alpha1.ListNodeMetricsRequest

func (r listNodeMetricsRequestExtended) validate() error {

	var (
		startTime *timestamp.Timestamp
		endTime   *timestamp.Timestamp
	)

	if r.TimeRange == nil {
		return errors.New("field \"time_range\" cannot be empty")
	}

	startTime = r.TimeRange.StartTime
	endTime = r.TimeRange.EndTime
	if startTime == nil || endTime == nil {
		return errors.New("field \"start_time\" and \"end_time\"  cannot be empty")
	}

	if startTime.Seconds+int64(startTime.Nanos) >= endTime.Seconds+int64(endTime.Nanos) {
		return errors.New("\"end_time\" must not be before \"start_time\"")
	}

	return nil
}

type podMetricExtended metric.PodMetric

func (p podMetricExtended) datahubPodMetric() datahub_v1alpha1.PodMetric {

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
		containerMetricExtended := containerMetricExtended(containerMetric)
		datahubContainerMetric := containerMetricExtended.datahubContainerMetric()
		datahubPodMetric.ContainerMetrics = append(datahubPodMetric.ContainerMetrics, &datahubContainerMetric)
	}

	return datahubPodMetric
}

type containerMetricExtended metric.ContainerMetric

func (c containerMetricExtended) NumberOfDatahubMetricDataNeededProducing() int {
	return 2
}

func (c containerMetricExtended) datahubContainerMetric() datahub_v1alpha1.ContainerMetric {

	var (
		metricDataChan = make(chan datahub_v1alpha1.MetricData)

		datahubContainerMetric datahub_v1alpha1.ContainerMetric
	)

	datahubContainerMetric = datahub_v1alpha1.ContainerMetric{
		Name: string(c.ContainerName),
	}

	go c.produceDatahubMetricDataFromSamples(datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE, c.CPUMetircs, metricDataChan)
	go c.produceDatahubMetricDataFromSamples(datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES, c.MemoryMetrics, metricDataChan)

	for i := 0; i < c.NumberOfDatahubMetricDataNeededProducing(); i++ {
		receivedMetricData := <-metricDataChan
		datahubContainerMetric.MetricData = append(datahubContainerMetric.MetricData, &receivedMetricData)
	}

	return datahubContainerMetric
}

func (c containerMetricExtended) produceDatahubMetricDataFromSamples(metricType datahub_v1alpha1.MetricType, samples []metric.Sample, metricDataChan chan<- datahub_v1alpha1.MetricData) {

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

	metricDataChan <- datahubMetricData
}

type nodeMetricExtended metric.NodeMetric

func (n nodeMetricExtended) NumberOfDatahubMetricDataNeededProducing() int {
	return 2
}

func (n nodeMetricExtended) datahubNodeMetric() datahub_v1alpha1.NodeMetric {

	var (
		metricDataChan = make(chan datahub_v1alpha1.MetricData)

		datahubNodeMetric datahub_v1alpha1.NodeMetric
	)

	datahubNodeMetric = datahub_v1alpha1.NodeMetric{
		Name: n.NodeName,
	}

	go n.produceDatahubMetricDataFromSamples(datahub_v1alpha1.MetricType_CPU_USAGE_SECONDS_PERCENTAGE, n.CPUUsageMetircs, metricDataChan)
	go n.produceDatahubMetricDataFromSamples(datahub_v1alpha1.MetricType_MEMORY_USAGE_BYTES, n.MemoryUsageMetrics, metricDataChan)

	for i := 0; i < n.NumberOfDatahubMetricDataNeededProducing(); i++ {
		receivedMetricData := <-metricDataChan
		datahubNodeMetric.MetricData = append(datahubNodeMetric.MetricData, &receivedMetricData)
	}

	return datahubNodeMetric
}

func (n nodeMetricExtended) produceDatahubMetricDataFromSamples(metricType datahub_v1alpha1.MetricType, samples []metric.Sample, metricDataChan chan<- datahub_v1alpha1.MetricData) {

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

	metricDataChan <- datahubMetricData
}
