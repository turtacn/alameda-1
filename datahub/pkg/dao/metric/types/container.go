package types

import (
	"fmt"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/datahub/pkg/metric"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	"sort"
)

type ContainerMetricSample struct {
	Namespace     metadata.NamespaceName
	PodName       metadata.PodName
	ContainerName metadata.ContainerName
	MetricType    metric.ContainerMetricType
	RateRange     int64
	Metrics       []metric.Sample
}

// ContainerMetric Metric model to represent one container metric
type ContainerMetric struct {
	Namespace     metadata.NamespaceName
	PodName       metadata.PodName
	ContainerName metadata.ContainerName
	RateRange     int64
	Metrics       map[metric.ContainerMetricType][]metric.Sample
}

// ContainersMetricMap Containers metric map
type ContainerMetricMap struct {
	MetricMap map[metadata.NamespacePodContainerName]*ContainerMetric
}

func NewContainerMetricSample() *ContainerMetricSample {
	metricSample := &ContainerMetricSample{}
	metricSample.Metrics = make([]metric.Sample, 0)
	return metricSample
}

func NewContainerMetric() *ContainerMetric {
	containerMetric := &ContainerMetric{}
	containerMetric.Metrics = make(map[metric.ContainerMetricType][]metric.Sample)
	return containerMetric
}

func NewContainerMetricMap() ContainerMetricMap {
	containerMetricMap := ContainerMetricMap{}
	containerMetricMap.MetricMap = make(map[metadata.NamespacePodContainerName]*ContainerMetric)
	return containerMetricMap
}

func (c *ContainerMetric) GetSamples(metricType metric.ContainerMetricType) *ContainerMetricSample {
	containerSample := NewContainerMetricSample()
	containerSample.Namespace = c.Namespace
	containerSample.PodName = c.PodName
	containerSample.ContainerName = c.ContainerName
	containerSample.MetricType = metricType
	containerSample.RateRange = c.RateRange

	if value, exist := c.Metrics[metricType]; exist {
		containerSample.Metrics = value
	}

	return containerSample
}

func (c *ContainerMetric) AddSample(metricType metric.ContainerMetricType, sample metric.Sample) {
	if _, exist := c.Metrics[metricType]; !exist {
		c.Metrics[metricType] = make([]metric.Sample, 0)
	}
	c.Metrics[metricType] = append(c.Metrics[metricType], sample)
}

// BuildPodMetric Build PodMetric consist of the receiver in ContainersMetricMap.
func (c *ContainerMetric) BuildPodMetric() *PodMetric {
	containerMetricMap := NewContainerMetricMap()
	containerMetricMap.MetricMap[c.NamespacePodContainerName()] = c

	return &PodMetric{
		Namespace:          c.Namespace,
		PodName:            c.PodName,
		RateRange:          c.RateRange,
		ContainerMetricMap: containerMetricMap,
	}
}

// NamespacePodContainerName Return identity of the container metric.
func (c *ContainerMetric) NamespacePodContainerName() metadata.NamespacePodContainerName {
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

// BuildPodsMetricMap Build PodsMetricMap base on current ContainersMetricMap
func (c *ContainerMetricMap) BuildPodsMetricMap() PodMetricMap {
	podsMetricMap := NewPodMetricMap()

	for _, containerMetric := range c.MetricMap {
		podsMetricMap.AddContainerMetric(containerMetric)
	}

	return podsMetricMap
}

// Merge Merge current ContainersMetricMap with input ContainersMetricMap
func (c *ContainerMetricMap) Merge(in ContainerMetricMap) {
	for namespacePodContainerName, containerMetric := range in.MetricMap {
		if existedContainerMetric, exist := c.MetricMap[namespacePodContainerName]; exist {
			for metricType, metrics := range containerMetric.Metrics {
				existedContainerMetric.Metrics[metricType] = append(existedContainerMetric.Metrics[metricType], metrics...)
			}
			c.MetricMap[namespacePodContainerName] = existedContainerMetric
		} else {
			c.MetricMap[namespacePodContainerName] = containerMetric
		}
	}
}
