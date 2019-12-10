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
		ObjectMeta: NewObjectMeta(&p.ObjectMeta),
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
		Name: string(c.ObjectMeta.Name),
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

	datahubNodeMetric = ApiMetrics.NodeMetric{}
	datahubNodeMetric.ObjectMeta = NewObjectMeta(&n.ObjectMeta)

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

type AppMetricExtended struct {
	DaoMetricTypes.AppMetric
}

func (n AppMetricExtended) ProduceMetrics() ApiMetrics.ApplicationMetric {
	var (
		m ApiMetrics.ApplicationMetric
	)

	m.ObjectMeta = &ApiResources.ObjectMeta{
		Namespace:   n.AppMetric.ObjectMeta.Namespace,
		Name:        n.AppMetric.ObjectMeta.Name,
		NodeName:    n.AppMetric.ObjectMeta.NodeName,
		ClusterName: n.AppMetric.ObjectMeta.ClusterName,
		Uid:         n.AppMetric.ObjectMeta.Uid,
	}
	m.MetricData = metricMapToDatahubMetricSlice(n.AppMetric.Metrics)
	return m
}

type ControllerMetricExtended struct {
	DaoMetricTypes.ControllerMetric
}

func (n ControllerMetricExtended) ProduceMetrics() ApiMetrics.ControllerMetric {
	var (
		m ApiMetrics.ControllerMetric
	)

	m.ObjectMeta = &ApiResources.ObjectMeta{
		Namespace:   n.ControllerMetric.ObjectMeta.Namespace,
		Name:        n.ControllerMetric.ObjectMeta.Name,
		NodeName:    n.ControllerMetric.ObjectMeta.NodeName,
		ClusterName: n.ControllerMetric.ObjectMeta.ClusterName,
		Uid:         n.ControllerMetric.ObjectMeta.Uid,
	}
	m.Kind = ApiResources.Kind(ApiResources.Kind_value[n.ControllerMetric.ObjectMeta.Kind])
	m.MetricData = metricMapToDatahubMetricSlice(n.ControllerMetric.Metrics)

	return m
}

type NamespaceMetricExtended struct {
	DaoMetricTypes.NamespaceMetric
}

func (n NamespaceMetricExtended) ProduceMetrics() ApiMetrics.NamespaceMetric {
	var (
		m ApiMetrics.NamespaceMetric
	)

	m.ObjectMeta = &ApiResources.ObjectMeta{
		Namespace:   n.NamespaceMetric.ObjectMeta.Namespace,
		Name:        n.NamespaceMetric.ObjectMeta.Name,
		NodeName:    n.NamespaceMetric.ObjectMeta.NodeName,
		ClusterName: n.NamespaceMetric.ObjectMeta.ClusterName,
		Uid:         n.NamespaceMetric.ObjectMeta.Uid,
	}
	m.MetricData = metricMapToDatahubMetricSlice(n.NamespaceMetric.Metrics)

	return m
}

type ClusterMetricExtended struct {
	DaoMetricTypes.ClusterMetric
}

func (n ClusterMetricExtended) ProduceMetrics() ApiMetrics.ClusterMetric {
	var (
		m ApiMetrics.ClusterMetric
	)

	m.ObjectMeta = &ApiResources.ObjectMeta{
		Namespace:   n.ClusterMetric.ObjectMeta.Namespace,
		Name:        n.ClusterMetric.ObjectMeta.Name,
		NodeName:    n.ClusterMetric.ObjectMeta.NodeName,
		ClusterName: n.ClusterMetric.ObjectMeta.ClusterName,
		Uid:         n.ClusterMetric.ObjectMeta.Uid,
	}
	m.MetricData = metricMapToDatahubMetricSlice(n.ClusterMetric.Metrics)

	return m
}
