package responses

import (
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	FormatEnum "github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiMetrics "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/metrics"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

type PodMetricExtended struct {
	*DaoMetricTypes.PodMetric
}

func (p *PodMetricExtended) ProduceMetrics() *ApiMetrics.PodMetric {
	var (
		datahubPodMetric ApiMetrics.PodMetric
	)

	datahubPodMetric = ApiMetrics.PodMetric{
		NamespacedName: &ApiResources.NamespacedName{
			Namespace: string(p.Namespace),
			Name:      string(p.PodName),
		},
	}

	for _, containerMetric := range p.ContainerMetricMap.MetricMap {
		containerMetricExtended := ContainerMetricExtended{containerMetric}
		datahubContainerMetric := containerMetricExtended.ProduceMetrics()
		datahubPodMetric.ContainerMetrics = append(datahubPodMetric.ContainerMetrics, datahubContainerMetric)
	}

	return &datahubPodMetric
}

type ContainerMetricExtended struct {
	*DaoMetricTypes.ContainerMetric
}

func (c *ContainerMetricExtended) ProduceMetrics() *ApiMetrics.ContainerMetric {
	var (
		metricDataChan  = make(chan ApiCommon.MetricData)
		numOfGoroutines = 0

		datahubContainerMetric ApiMetrics.ContainerMetric
	)

	datahubContainerMetric = ApiMetrics.ContainerMetric{
		Name: string(c.ContainerName),
	}

	for metricType, samples := range c.Metrics {
		if datahubMetricType, exist := FormatEnum.TypeToDatahubMetricType[metricType]; exist {
			numOfGoroutines++
			go produceMetricDataFromSamples(datahubMetricType, samples, metricDataChan)
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

func (n *NodeMetricExtended) ProduceMetrics() *ApiMetrics.NodeMetric {
	var (
		metricDataChan  = make(chan ApiCommon.MetricData)
		numOfGoroutines = 0

		datahubNodeMetric ApiMetrics.NodeMetric
	)

	datahubNodeMetric = ApiMetrics.NodeMetric{
		Name: n.NodeName,
	}

	for metricType, samples := range n.Metrics {
		if datahubMetricType, exist := FormatEnum.TypeToDatahubMetricType[metricType]; exist {
			numOfGoroutines++
			go produceMetricDataFromSamples(datahubMetricType, samples, metricDataChan)
		}
	}

	for i := 0; i < numOfGoroutines; i++ {
		receivedMetricData := <-metricDataChan
		datahubNodeMetric.MetricData = append(datahubNodeMetric.MetricData, &receivedMetricData)
	}

	return &datahubNodeMetric
}
