package types

import (
	"fmt"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
)

// PodMetricsDAO DAO interface of pod metric data.
type PodMetricsDAO interface {
	CreateMetrics(PodMetricMap) error
	ListMetrics(ListPodMetricsRequest) (PodMetricMap, error)
}

// PodMetric Metric model to represent one pod's metric
type PodMetric struct {
	Namespace          metadata.NamespaceName
	PodName            metadata.PodName
	RateRange          int64
	ContainerMetricMap ContainerMetricMap
}

// PodsMetricMap Pods' metric map
type PodMetricMap struct {
	MetricMap map[metadata.NamespacePodName]*PodMetric
}

// ListPodMetricsRequest Argument of method ListPodMetrics
type ListPodMetricsRequest struct {
	DBCommon.QueryCondition
	Namespace metadata.NamespaceName
	PodName   metadata.PodName
	RateRange int64
}

func NewPodMetric() *PodMetric {
	nodeMetric := &PodMetric{}
	nodeMetric.ContainerMetricMap = NewContainerMetricMap()
	return nodeMetric
}

func NewPodMetricMap() PodMetricMap {
	podMetricMap := PodMetricMap{}
	podMetricMap.MetricMap = make(map[metadata.NamespacePodName]*PodMetric)
	return podMetricMap
}

// NamespacePodName Return identity of the pod metric
func (p *PodMetric) NamespacePodName() metadata.NamespacePodName {
	return metadata.NamespacePodName(fmt.Sprintf("%s/%s", p.Namespace, p.PodName))
}

// Merge Merge current PodMetric with input PodMetric
func (p *PodMetric) Merge(in *PodMetric) {
	p.ContainerMetricMap.Merge(in.ContainerMetricMap)
}

// SortByTimestamp Sort each container metric's content
func (p *PodMetric) SortByTimestamp(order DBCommon.Order) {
	for _, containerMetric := range p.ContainerMetricMap.MetricMap {
		containerMetric.SortByTimestamp(order)
	}
}

// Limit Slicing each container metric content
func (p *PodMetric) Limit(limit int) {
	for _, containerMetric := range p.ContainerMetricMap.MetricMap {
		containerMetric.Limit(limit)
	}
}

// AddContainerMetric Add container metric into PodsMetricMap
func (p *PodMetricMap) AddContainerMetric(c *ContainerMetric) {
	podMetric := c.BuildPodMetric()
	namespacePodName := podMetric.NamespacePodName()
	if existedPodMetric, exist := p.MetricMap[namespacePodName]; exist {
		existedPodMetric.Merge(podMetric)
	} else {
		p.MetricMap[namespacePodName] = podMetric
	}
}

// SortByTimestamp Sort each pod metric's content
func (p *PodMetricMap) SortByTimestamp(order DBCommon.Order) {
	for _, podMetric := range p.MetricMap {
		podMetric.SortByTimestamp(order)
	}
}

// Limit Slicing each pod metric content
func (p *PodMetricMap) Limit(limit int) {
	for _, podMetric := range p.MetricMap {
		podMetric.Limit(limit)
	}
}
