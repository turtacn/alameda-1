package types

import (
	"context"
	"sort"

	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
)

// ClusterMetricsDAO DAO interface of namespace metric data.
type ClusterMetricsDAO interface {
	CreateMetrics(context.Context, ClusterMetricMap) error
	ListMetrics(context.Context, ListClusterMetricsRequest) (ClusterMetricMap, error)
}

type ClusterMetricSample struct {
	ObjectMeta metadata.ObjectMeta
	MetricType enumconv.MetricType
	Metrics    []types.Sample
}

// ClusterMetric Metric model to represent one namespace metric
type ClusterMetric struct {
	ObjectMeta metadata.ObjectMeta
	Metrics    map[enumconv.MetricType][]types.Sample
}

// ClusterMetricMap Clusters' metric map
type ClusterMetricMap struct {
	MetricMap map[metadata.ObjectMeta]*ClusterMetric
}

// ListClusterMetricsRequest Argument of method ListClusterMetrics
type ListClusterMetricsRequest struct {
	common.QueryCondition
	ObjectMetas []metadata.ObjectMeta
}

func NewClusterMetric() *ClusterMetric {
	ClusterMetric := &ClusterMetric{}
	ClusterMetric.Metrics = make(map[enumconv.MetricType][]types.Sample)
	return ClusterMetric
}

func (n *ClusterMetric) AddSample(metricType enumconv.MetricType, sample types.Sample) {
	if n.Metrics == nil {
		n.Metrics = make(map[enumconv.MetricType][]types.Sample)
	}
	if _, exist := n.Metrics[metricType]; !exist {
		n.Metrics[metricType] = make([]types.Sample, 0)
	}
	n.Metrics[metricType] = append(n.Metrics[metricType], sample)
}

func (c *ClusterMetric) GetSamples(metricType enumconv.MetricType) ClusterMetricSample {
	if c.Metrics == nil {
		c.Metrics = make(map[enumconv.MetricType][]types.Sample)
	}
	return ClusterMetricSample{
		ObjectMeta: c.ObjectMeta,
		MetricType: metricType,
		Metrics:    c.Metrics[metricType],
	}
}

// Merge Merge current ClusterMetric with input ClusterMetric
func (n *ClusterMetric) Merge(in *ClusterMetric) {
	if n.Metrics == nil {
		n.Metrics = make(map[enumconv.MetricType][]types.Sample)
	}
	for metricType, metrics := range in.Metrics {
		n.Metrics[metricType] = append(n.Metrics[metricType], metrics...)
	}
}

// SortByTimestamp Sort each metric samples by timestamp in input order
func (n *ClusterMetric) SortByTimestamp(order common.Order) {
	for _, samples := range n.Metrics {
		if order == common.Asc {
			sort.Sort(types.SamplesByAscTimestamp(samples))
		} else {
			sort.Sort(types.SamplesByDescTimestamp(samples))
		}
	}
}

// Limit Slicing each metric samples element
func (n *ClusterMetric) Limit(limit int) {

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

func NewClusterMetricMap() ClusterMetricMap {
	ClusterMetricMap := ClusterMetricMap{}
	ClusterMetricMap.MetricMap = make(map[metadata.ObjectMeta]*ClusterMetric)
	return ClusterMetricMap
}

// AddClusterMetric Add namespace metric into ClustersMetricMap
func (n *ClusterMetricMap) AddClusterMetric(m *ClusterMetric) {
	if n.MetricMap == nil {
		n.MetricMap = make(map[metadata.ObjectMeta]*ClusterMetric)
	}
	if existClusterMetric, exist := n.MetricMap[m.ObjectMeta]; exist {
		existClusterMetric.Merge(m)
	} else {
		n.MetricMap[m.ObjectMeta] = m
	}
}

func (n *ClusterMetricMap) GetClusterMetric(o metadata.ObjectMeta) ClusterMetric {
	if n.MetricMap == nil {
		n.MetricMap = make(map[metadata.ObjectMeta]*ClusterMetric)
	}
	m, exist := n.MetricMap[o]
	if !exist {
		m = &ClusterMetric{}
	}
	return *m
}

func (c *ClusterMetricMap) GetSamples(metricType enumconv.MetricType) []ClusterMetricSample {
	custerMetricSample := make([]ClusterMetricSample, 0, len(c.MetricMap))
	for _, metric := range c.MetricMap {
		if metric == nil {
			continue
		}
		custerMetricSample = append(custerMetricSample, metric.GetSamples(metricType))
	}

	return custerMetricSample
}

// SortByTimestamp Sort each namespace metric's content
func (n *ClusterMetricMap) SortByTimestamp(order common.Order) {
	for _, ClusterMetric := range n.MetricMap {
		ClusterMetric.SortByTimestamp(order)
	}
}

// Limit Limit each namespace metric's content
func (n *ClusterMetricMap) Limit(limit int) {
	for _, ClusterMetric := range n.MetricMap {
		ClusterMetric.Limit(limit)
	}
}
