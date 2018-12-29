package metric

import (
	"fmt"
	"time"
)

type MetricsDAO interface {
	ListPodMetrics(ListPodMetricsRequest) (PodsMetricMap, error)
}

type NamespaceName string
type PodName string
type ContainerName string
type NamespacePodName string
type NamespacePodContainerName string

type ListPodMetricsRequest struct {
	Namespace string
	PodName   string
	StartTime time.Time
	EndTime   time.Time
}

type Sample struct {
	Timestamp time.Time
	Value     string
}

type ContainerMetric struct {
	Namespace     NamespaceName
	PodName       PodName
	ContainerName ContainerName
	CPUMetircs    []Sample
	MemoryMetrics []Sample
}

func (c ContainerMetric) BuildPodMetric() PodMetric {

	containersMetricMap := ContainersMetricMap{}
	containersMetricMap[c.NamespacePodContainerName()] = c

	return PodMetric{
		Namespace:           c.Namespace,
		PodName:             c.PodName,
		ContainersMetricMap: containersMetricMap,
	}
}

func (c ContainerMetric) NamespacePodContainerName() NamespacePodContainerName {
	return NamespacePodContainerName(fmt.Sprintf("%s/%s/%s", c.Namespace, c.PodName, c.ContainerName))
}

type ContainersMetricMap map[NamespacePodContainerName]ContainerMetric

// PodsMetricMap Build PodsMetricMap base on current ContainersMetricMap
func (c ContainersMetricMap) BuildPodsMetricMap() PodsMetricMap {

	var (
		podsMetricMap = &PodsMetricMap{}
	)

	for _, containerMetric := range c {
		podsMetricMap.AddContainerMetric(containerMetric)
	}

	return *podsMetricMap
}

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

type PodMetric struct {
	Namespace           NamespaceName
	PodName             PodName
	ContainersMetricMap ContainersMetricMap
}

func (p PodMetric) NamespacePodName() NamespacePodName {
	return NamespacePodName(fmt.Sprintf("%s/%s", p.Namespace, p.PodName))
}

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

type PodsMetricMap map[NamespacePodName]PodMetric

func (p *PodsMetricMap) AddContainerMetric(c ContainerMetric) {

	podMetric := c.BuildPodMetric()
	namespacePodName := podMetric.NamespacePodName()
	if existedPodMetric, exist := (*p)[namespacePodName]; exist {
		(*p)[namespacePodName] = existedPodMetric.Merge(podMetric)
	} else {
		(*p)[namespacePodName] = podMetric
	}
}
