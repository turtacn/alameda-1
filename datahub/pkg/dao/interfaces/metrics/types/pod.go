package types

import (
	"context"
	"fmt"

	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
)

// PodMetricsDAO DAO interface of pod metric data.
type PodMetricsDAO interface {
	CreateMetrics(context.Context, PodMetricMap) error
	ListMetrics(context.Context, ListPodMetricsRequest) (PodMetricMap, error)
}

// PodMetric Metric model to represent one pod's metric
type PodMetric struct {
	ObjectMeta         metadata.ObjectMeta
	RateRange          int64
	ContainerMetricMap ContainerMetricMap
}

// PodMetricMap Pods' metric map
type PodMetricMap struct {
	MetricMap map[metadata.ObjectMeta]*PodMetric
}

// ListPodMetricsRequest Argument of method ListPodMetrics
type ListPodMetricsRequest struct {
	common.QueryCondition
	ObjectMetas []*metadata.ObjectMeta
	RateRange   int64
}

func NewPodMetric() *PodMetric {
	nodeMetric := &PodMetric{}
	nodeMetric.ContainerMetricMap = NewContainerMetricMap()
	return nodeMetric
}

func NewPodMetricMap() PodMetricMap {
	podMetricMap := PodMetricMap{}
	podMetricMap.MetricMap = make(map[metadata.ObjectMeta]*PodMetric)
	return podMetricMap
}

func NewListPodMetricsRequest() ListPodMetricsRequest {
	request := ListPodMetricsRequest{}
	request.ObjectMetas = make([]*metadata.ObjectMeta, 0)
	return request
}

// NamespacePodName Return identity of the pod metric
func (p *PodMetric) NamespacePodName() metadata.NamespacePodName {
	return metadata.NamespacePodName(fmt.Sprintf("%s/%s", p.ObjectMeta.Namespace, p.ObjectMeta.Name))
}

// Merge Merge current PodMetric with input PodMetric
func (p *PodMetric) Merge(in *PodMetric) {
	if p.ContainerMetricMap.MetricMap == nil {
		p.ContainerMetricMap.MetricMap = make(map[ContainerMeta]*ContainerMetric)
	}
	for _, containerMetric := range in.ContainerMetricMap.MetricMap {
		p.ContainerMetricMap.AddContainerMetric(containerMetric)
	}
}

// SortByTimestamp Sort each container metric's content
func (p *PodMetric) SortByTimestamp(order common.Order) {
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

func (p *PodMetricMap) AddPodMetric(podMetric *PodMetric) {
	if p.MetricMap == nil {
		p.MetricMap = make(map[metadata.ObjectMeta]*PodMetric)
	}
	if existedPodMetric, exist := p.MetricMap[podMetric.ObjectMeta]; exist {
		existedPodMetric.Merge(podMetric)
	} else {
		p.MetricMap[podMetric.ObjectMeta] = podMetric
	}
}

// AddContainerMetric Add container metric into PodsMetricMap
func (p *PodMetricMap) AddContainerMetric(c *ContainerMetric) {
	// TODO
	if p.MetricMap == nil {
		p.MetricMap = make(map[metadata.ObjectMeta]*PodMetric)
	}
	podMetric := c.BuildPodMetric()
	if existedPodMetric, exist := p.MetricMap[podMetric.ObjectMeta]; exist {
		existedPodMetric.Merge(podMetric)
	} else {
		p.MetricMap[podMetric.ObjectMeta] = podMetric
	}
}

// SortByTimestamp Sort each pod metric's content
func (p *PodMetricMap) SortByTimestamp(order common.Order) {
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
