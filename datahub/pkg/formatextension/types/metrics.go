package types

import (
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/metric/types"
	Metric "github.com/containers-ai/alameda/datahub/pkg/metric"
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

type PodMetricExtended struct {
	*DaoMetricTypes.PodMetric
}

func (p *PodMetricExtended) ProduceMetrics() *DatahubV1alpha1.PodMetric {
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
		containerMetricExtended := ContainerMetricExtended{containerMetric}
		datahubContainerMetric := containerMetricExtended.ProduceMetrics()
		datahubPodMetric.ContainerMetrics = append(datahubPodMetric.ContainerMetrics, datahubContainerMetric)
	}

	return &datahubPodMetric
}

type ContainerMetricExtended struct {
	*DaoMetricTypes.ContainerMetric
}

func (c *ContainerMetricExtended) ProduceMetrics() *DatahubV1alpha1.ContainerMetric {
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

type NodeMetricExtended struct {
	*DaoMetricTypes.NodeMetric
}

func (n *NodeMetricExtended) ProduceMetrics() *DatahubV1alpha1.NodeMetric {
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
