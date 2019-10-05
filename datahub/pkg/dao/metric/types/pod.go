package types

import (
	"fmt"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/datahub/pkg/metric"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	"sort"
)

// PodMetricsDAO DAO interface of pod metric data.
type PodMetricsDAO interface {
	ListMetrics(ListPodMetricsRequest) (PodsMetricMap, error)
}

// ListPodMetricsRequest Argument of method ListPodMetrics
type ListPodMetricsRequest struct {
	Namespace metadata.NamespaceName
	PodName   metadata.PodName
	DBCommon.QueryCondition
}

// ContainerMetric Metric model to represent one container metric
type ContainerMetric struct {
	Namespace     metadata.NamespaceName
	PodName       metadata.PodName
	ContainerName metadata.ContainerName
	Metrics       map[metric.ContainerMetricType][]metric.Sample
}

// BuildPodMetric Build PodMetric consist of the receiver in ContainersMetricMap.
func (c *ContainerMetric) BuildPodMetric() *PodMetric {

	containersMetricMap := ContainersMetricMap{}
	containersMetricMap[c.NamespacePodContainerName()] = c

	return &PodMetric{
		Namespace:           c.Namespace,
		PodName:             c.PodName,
		ContainersMetricMap: &containersMetricMap,
	}
}

// NamespacePodContainerName Return identity of the container metric.
func (c ContainerMetric) NamespacePodContainerName() metadata.NamespacePodContainerName {
	return metadata.NamespacePodContainerName(fmt.Sprintf("%s/%s/%s", c.Namespace, c.PodName, c.ContainerName))
}

// SortByTimestamp Sort each metric samples by timestamp in input order
func (c *ContainerMetric) SortByTimestamp(order DBCommon.Order) {

	for _, samples := range c.Metrics {
		if order == DBCommon.Asc {
			sort.Sort(metric.SamplesByAscTimestamp(samples))
		} else {
			sort.Sort(metric.SamplesByDescTimestamp(samples))
		}
	}
}

// Limit Slicing each metric samples element
func (c *ContainerMetric) Limit(limit int) {

	if limit == 0 {
		return
	}

	for metricType, samples := range c.Metrics {
		c.Metrics[metricType] = samples[:limit]
	}
}

// ContainersMetricMap Containers metric map
type ContainersMetricMap map[metadata.NamespacePodContainerName]*ContainerMetric

// BuildPodsMetricMap Build PodsMetricMap base on current ContainersMetricMap
func (c ContainersMetricMap) BuildPodsMetricMap() *PodsMetricMap {

	var (
		podsMetricMap = &PodsMetricMap{}
	)

	for _, containerMetric := range c {
		podsMetricMap.AddContainerMetric(containerMetric)
	}

	return podsMetricMap
}

// Merge Merge current ContainersMetricMap with input ContainersMetricMap
func (c *ContainersMetricMap) Merge(in *ContainersMetricMap) {

	for namespacePodContainerName, containerMetric := range *in {
		if existedContainerMetric, exist := (*c)[namespacePodContainerName]; exist {
			for metricType, metrics := range containerMetric.Metrics {
				existedContainerMetric.Metrics[metricType] = append(existedContainerMetric.Metrics[metricType], metrics...)
			}
			(*c)[namespacePodContainerName] = existedContainerMetric
		} else {
			(*c)[namespacePodContainerName] = containerMetric
		}
	}
}

// PodMetric Metric model to represent one pod's metric
type PodMetric struct {
	Namespace           metadata.NamespaceName
	PodName             metadata.PodName
	ContainersMetricMap *ContainersMetricMap
}

// NamespacePodName Return identity of the pod metric
func (p PodMetric) NamespacePodName() metadata.NamespacePodName {
	return metadata.NamespacePodName(fmt.Sprintf("%s/%s", p.Namespace, p.PodName))
}

// Merge Merge current PodMetric with input PodMetric
func (p *PodMetric) Merge(in *PodMetric) {
	p.ContainersMetricMap.Merge(in.ContainersMetricMap)
}

// SortByTimestamp Sort each container metric's content
func (p *PodMetric) SortByTimestamp(order DBCommon.Order) {

	for _, containerMetric := range *p.ContainersMetricMap {
		containerMetric.SortByTimestamp(order)
	}
}

// Limit Slicing each container metric content
func (p *PodMetric) Limit(limit int) {

	for _, containerMetric := range *p.ContainersMetricMap {
		containerMetric.Limit(limit)
	}
}

// PodsMetricMap Pods' metric map
type PodsMetricMap map[metadata.NamespacePodName]*PodMetric

// AddContainerMetric Add container metric into PodsMetricMap
func (p *PodsMetricMap) AddContainerMetric(c *ContainerMetric) {

	podMetric := c.BuildPodMetric()
	namespacePodName := podMetric.NamespacePodName()
	if existedPodMetric, exist := (*p)[namespacePodName]; exist {
		existedPodMetric.Merge(podMetric)
	} else {
		(*p)[namespacePodName] = podMetric
	}
}

// SortByTimestamp Sort each pod metric's content
func (p *PodsMetricMap) SortByTimestamp(order DBCommon.Order) {

	for _, podMetric := range *p {
		podMetric.SortByTimestamp(order)
	}
}

// Limit Slicing each pod metric content
func (p *PodsMetricMap) Limit(limit int) {

	for _, podMetric := range *p {
		podMetric.Limit(limit)
	}
}
