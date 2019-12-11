package types

import (
	"sort"

	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
)

type ContainerMeta struct {
	metadata.ObjectMeta
	PodName string
}

type ContainerMetricSample struct {
	ObjectMeta ContainerMeta
	MetricType enumconv.MetricType
	RateRange  int64
	Metrics    []types.Sample
}

// ContainerMetric Metric model to represent one container metric
type ContainerMetric struct {
	ObjectMeta ContainerMeta
	RateRange  int64
	Metrics    map[enumconv.MetricType][]types.Sample
}

// ContainerMetricMap Containers metric map
type ContainerMetricMap struct {
	MetricMap map[ContainerMeta]*ContainerMetric
}

func NewContainerMetricSample() *ContainerMetricSample {
	metricSample := &ContainerMetricSample{}
	metricSample.Metrics = make([]types.Sample, 0)
	return metricSample
}

func NewContainerMetric() *ContainerMetric {
	containerMetric := &ContainerMetric{}
	containerMetric.Metrics = make(map[enumconv.MetricType][]types.Sample)
	return containerMetric
}

func NewContainerMetricMap() ContainerMetricMap {
	containerMetricMap := ContainerMetricMap{}
	containerMetricMap.MetricMap = make(map[ContainerMeta]*ContainerMetric)
	return containerMetricMap
}

func (c *ContainerMetric) GetSamples(metricType enumconv.MetricType) *ContainerMetricSample {
	containerSample := NewContainerMetricSample()
	containerSample.ObjectMeta = c.ObjectMeta
	containerSample.MetricType = metricType
	containerSample.RateRange = c.RateRange

	if c.Metrics == nil {
		c.Metrics = make(map[enumconv.MetricType][]types.Sample)
	}
	if value, exist := c.Metrics[metricType]; exist {
		containerSample.Metrics = value
	}

	return containerSample
}

func (c *ContainerMetric) AddSample(metricType enumconv.MetricType, sample types.Sample) {
	if c.Metrics == nil {
		c.Metrics = make(map[enumconv.MetricType][]types.Sample)
	}
	if _, exist := c.Metrics[metricType]; !exist {
		c.Metrics[metricType] = make([]types.Sample, 0)
	}
	c.Metrics[metricType] = append(c.Metrics[metricType], sample)
}

// Merge Merge current ContainersMetricMap with input ContainersMetricMap
func (c *ContainerMetric) Merge(in *ContainerMetric) {
	if c.Metrics == nil {
		c.Metrics = make(map[enumconv.MetricType][]types.Sample)
	}
	for metricType, containerMetric := range in.Metrics {
		if _, exist := c.Metrics[metricType]; exist {
			c.Metrics[metricType] = append(c.Metrics[metricType], containerMetric...)
		} else {
			c.Metrics[metricType] = containerMetric
		}
	}
}

// BuildPodMetric Build PodMetric consist of the receiver in ContainersMetricMap.
func (c *ContainerMetric) BuildPodMetric() *PodMetric {
	containerMetricMap := NewContainerMetricMap()
	containerMetricMap.MetricMap[c.ObjectMeta] = c

	return &PodMetric{
		ObjectMeta: metadata.ObjectMeta{
			Name:        c.ObjectMeta.PodName,
			Namespace:   c.ObjectMeta.Namespace,
			ClusterName: c.ObjectMeta.ClusterName,
			NodeName:    c.ObjectMeta.NodeName,
		},
		RateRange:          c.RateRange,
		ContainerMetricMap: containerMetricMap,
	}
}

// SortByTimestamp Sort each metric samples by timestamp in input order
func (c *ContainerMetric) SortByTimestamp(order common.Order) {
	for _, samples := range c.Metrics {
		if order == common.Asc {
			sort.Sort(types.SamplesByAscTimestamp(samples))
		} else {
			sort.Sort(types.SamplesByDescTimestamp(samples))
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

func (c *ContainerMetricMap) AddContainerMetric(containerMetric *ContainerMetric) {
	if c.MetricMap == nil {
		c.MetricMap = make(map[ContainerMeta]*ContainerMetric)
	}
	if existContainerMetric, exist := c.MetricMap[containerMetric.ObjectMeta]; exist {
		existContainerMetric.Merge(containerMetric)
	} else {
		c.MetricMap[containerMetric.ObjectMeta] = containerMetric
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
