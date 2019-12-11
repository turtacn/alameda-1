package types

import (
	"context"
	"sort"

	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
)

// NamespaceMetricsDAO DAO interface of namespace metric data.
type NamespaceMetricsDAO interface {
	CreateMetrics(context.Context, NamespaceMetricMap) error
	ListMetrics(context.Context, ListNamespaceMetricsRequest) (NamespaceMetricMap, error)
}

type NamespaceMetricSample struct {
	ObjectMeta metadata.ObjectMeta
	MetricType enumconv.MetricType
	Metrics    []types.Sample
}

// NamespaceMetric Metric model to represent one namespace metric
type NamespaceMetric struct {
	ObjectMeta metadata.ObjectMeta
	Metrics    map[enumconv.MetricType][]types.Sample
}

// NamespaceMetricMap Namespaces' metric map
type NamespaceMetricMap struct {
	MetricMap map[metadata.ObjectMeta]*NamespaceMetric
}

// ListNamespaceMetricsRequest Argument of method ListNamespaceMetrics
type ListNamespaceMetricsRequest struct {
	common.QueryCondition
	ObjectMetas []metadata.ObjectMeta
}

func NewNamespaceMetric() *NamespaceMetric {
	NamespaceMetric := &NamespaceMetric{}
	NamespaceMetric.Metrics = make(map[enumconv.MetricType][]types.Sample)
	return NamespaceMetric
}

func (n *NamespaceMetric) AddSample(metricType enumconv.MetricType, sample types.Sample) {
	if n.Metrics == nil {
		n.Metrics = make(map[enumconv.MetricType][]types.Sample)
	}
	if _, exist := n.Metrics[metricType]; !exist {
		n.Metrics[metricType] = make([]types.Sample, 0)
	}
	n.Metrics[metricType] = append(n.Metrics[metricType], sample)
}

func (c *NamespaceMetric) GetSamples(metricType enumconv.MetricType) NamespaceMetricSample {
	if c.Metrics == nil {
		c.Metrics = make(map[enumconv.MetricType][]types.Sample)
	}
	return NamespaceMetricSample{
		ObjectMeta: c.ObjectMeta,
		MetricType: metricType,
		Metrics:    c.Metrics[metricType],
	}
}

// Merge Merge current NamespaceMetric with input NamespaceMetric
func (n *NamespaceMetric) Merge(in *NamespaceMetric) {
	if n.Metrics == nil {
		n.Metrics = make(map[enumconv.MetricType][]types.Sample)
	}
	for metricType, metrics := range in.Metrics {
		n.Metrics[metricType] = append(n.Metrics[metricType], metrics...)
	}
}

// SortByTimestamp Sort each metric samples by timestamp in input order
func (n *NamespaceMetric) SortByTimestamp(order common.Order) {
	for _, samples := range n.Metrics {
		if order == common.Asc {
			sort.Sort(types.SamplesByAscTimestamp(samples))
		} else {
			sort.Sort(types.SamplesByDescTimestamp(samples))
		}
	}
}

// Limit Slicing each metric samples element
func (n *NamespaceMetric) Limit(limit int) {

	if limit == 0 {
		return
	}

	for metricType, samples := range n.Metrics {
		if len(samples) < limit {
			continue
		}
		n.Metrics[metricType] = samples[:limit]
	}
}

func NewNamespaceMetricMap() NamespaceMetricMap {
	NamespaceMetricMap := NamespaceMetricMap{}
	NamespaceMetricMap.MetricMap = make(map[metadata.ObjectMeta]*NamespaceMetric)
	return NamespaceMetricMap
}

// AddNamespaceMetric Add namespace metric into NamespacesMetricMap
func (n *NamespaceMetricMap) AddNamespaceMetric(m *NamespaceMetric) {
	if n.MetricMap == nil {
		n.MetricMap = make(map[metadata.ObjectMeta]*NamespaceMetric)
	}
	if existNamespaceMetric, exist := n.MetricMap[m.ObjectMeta]; exist {
		existNamespaceMetric.Merge(m)
	} else {
		n.MetricMap[m.ObjectMeta] = m
	}
}

func (c *NamespaceMetricMap) GetSamples(metricType enumconv.MetricType) []NamespaceMetricSample {
	namespaceMetricSamples := make([]NamespaceMetricSample, 0, len(c.MetricMap))
	for _, metric := range c.MetricMap {
		if metric == nil {
			continue
		}
		namespaceMetricSamples = append(namespaceMetricSamples, metric.GetSamples(metricType))
	}
	return namespaceMetricSamples
}

// SortByTimestamp Sort each namespace metric's content
func (n *NamespaceMetricMap) SortByTimestamp(order common.Order) {
	for _, NamespaceMetric := range n.MetricMap {
		NamespaceMetric.SortByTimestamp(order)
	}
}

// Limit Limit each namespace metric's content
func (n *NamespaceMetricMap) Limit(limit int) {
	for _, NamespaceMetric := range n.MetricMap {
		NamespaceMetric.Limit(limit)
	}
}
