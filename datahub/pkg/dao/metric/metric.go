package metric

import (
	"fmt"
	"time"
)

// MetricsDAO DAO interface of metric data.
type MetricsDAO interface {
	ListPodMetrics(ListPodMetricsRequest) (PodsMetricMap, error)
	ListNodesMetric(ListNodeMetricsRequest) (NodesMetricMap, error)
}

// NamespaceName Type alias
type NamespaceName = string

// PodName Type alias
type PodName = string

// ContainerName Type alias
type ContainerName = string

// NodeName Type alias
type NodeName = string

// NamespacePodName Type alias
type NamespacePodName = string

// NamespacePodContainerName Type alias
type NamespacePodContainerName = string

// ListPodMetricsRequest Argument of method ListPodMetrics
type ListPodMetricsRequest struct {
	Namespace string
	PodName   string
	StartTime time.Time
	EndTime   time.Time
}

// ListNodeMetricsRequest Argument of method ListNodeMetrics
type ListNodeMetricsRequest struct {
	NodeNames []NodeName
	StartTime time.Time
	EndTime   time.Time
}

// Sample Data struct representing timestamp and metric value of metric data point
type Sample struct {
	Timestamp time.Time
	Value     string
}

// ContainerMetric Metric model to represent one container metric
type ContainerMetric struct {
	Namespace     NamespaceName
	PodName       PodName
	ContainerName ContainerName
	CPUMetircs    []Sample
	MemoryMetrics []Sample
}

// BuildPodMetric Build PodMetric consist of the receiver in ContainersMetricMap.
func (c ContainerMetric) BuildPodMetric() PodMetric {

	containersMetricMap := ContainersMetricMap{}
	containersMetricMap[c.NamespacePodContainerName()] = c

	return PodMetric{
		Namespace:           c.Namespace,
		PodName:             c.PodName,
		ContainersMetricMap: containersMetricMap,
	}
}

// NamespacePodContainerName Return identity of the container metric.
func (c ContainerMetric) NamespacePodContainerName() NamespacePodContainerName {
	return NamespacePodContainerName(fmt.Sprintf("%s/%s/%s", c.Namespace, c.PodName, c.ContainerName))
}

// ContainersMetricMap Containers metric map
type ContainersMetricMap map[NamespacePodContainerName]ContainerMetric

// BuildPodsMetricMap Build PodsMetricMap base on current ContainersMetricMap
func (c ContainersMetricMap) BuildPodsMetricMap() PodsMetricMap {

	var (
		podsMetricMap = &PodsMetricMap{}
	)

	for _, containerMetric := range c {
		podsMetricMap.AddContainerMetric(containerMetric)
	}

	return *podsMetricMap
}

// Merge Merge current ContainersMetricMap with input ContainersMetricMap
func (c ContainersMetricMap) Merge(in ContainersMetricMap) ContainersMetricMap {

	for namespacePodContainerName, containerMetric := range in {
		if existedContainerMetric, exist := c[namespacePodContainerName]; exist {
			existedContainerMetric.CPUMetircs = append(existedContainerMetric.CPUMetircs, containerMetric.CPUMetircs...)
			existedContainerMetric.MemoryMetrics = append(existedContainerMetric.MemoryMetrics, containerMetric.MemoryMetrics...)
			c[namespacePodContainerName] = existedContainerMetric
		} else {
			c[namespacePodContainerName] = containerMetric
		}
	}

	return c
}

// PodMetric Metric model to represent one pod's metric
type PodMetric struct {
	Namespace           NamespaceName
	PodName             PodName
	ContainersMetricMap ContainersMetricMap
}

// NamespacePodName Return identity of the pod metric
func (p PodMetric) NamespacePodName() NamespacePodName {
	return NamespacePodName(fmt.Sprintf("%s/%s", p.Namespace, p.PodName))
}

// Merge Merge current PodMetric with input PodMetric
func (p PodMetric) Merge(in PodMetric) PodMetric {

	var (
		currentContainerMetircMap   = p.ContainersMetricMap
		mergeWithContainerMetircMap = in.ContainersMetricMap
		newPodMetric                = PodMetric{
			Namespace:           p.Namespace,
			PodName:             p.PodName,
			ContainersMetricMap: currentContainerMetircMap.Merge(mergeWithContainerMetircMap),
		}
	)

	return newPodMetric
}

// PodsMetricMap Pods' metric map
type PodsMetricMap map[NamespacePodName]PodMetric

// AddContainerMetric Add container metric into PodsMetricMap
func (p *PodsMetricMap) AddContainerMetric(c ContainerMetric) {

	podMetric := c.BuildPodMetric()
	namespacePodName := podMetric.NamespacePodName()
	if existedPodMetric, exist := (*p)[namespacePodName]; exist {
		(*p)[namespacePodName] = existedPodMetric.Merge(podMetric)
	} else {
		(*p)[namespacePodName] = podMetric
	}
}

// NodeMetric Metric model to represent one node metric
type NodeMetric struct {
	NodeName               NodeName
	CPUUsageMetircs        []Sample
	MemoryTotalMetrics     []Sample
	MemoryAvailableMetrics []Sample
	MemoryUsageMetrics     []Sample
}

// Merge Merge current NodeMetric with input NodeMetric
func (n NodeMetric) Merge(in NodeMetric) NodeMetric {

	var (
		newNodeMetirc = NodeMetric{
			NodeName:               n.NodeName,
			CPUUsageMetircs:        append(n.CPUUsageMetircs, in.CPUUsageMetircs...),
			MemoryTotalMetrics:     append(n.MemoryTotalMetrics, in.MemoryTotalMetrics...),
			MemoryAvailableMetrics: append(n.MemoryAvailableMetrics, in.MemoryAvailableMetrics...),
			MemoryUsageMetrics:     append(n.MemoryUsageMetrics, in.MemoryUsageMetrics...),
		}
	)

	return newNodeMetirc
}

// NodesMetricMap Nodes' metric map
type NodesMetricMap map[NodeName]NodeMetric

// AddNodeMetric Add node metric into NodesMetricMap
func (n *NodesMetricMap) AddNodeMetric(nodeMetric NodeMetric) {

	nodeName := nodeMetric.NodeName
	if existNodeMetric, exist := (*n)[nodeName]; exist {
		(*n)[nodeName] = existNodeMetric.Merge(nodeMetric)
	} else {
		(*n)[nodeName] = nodeMetric
	}
}
