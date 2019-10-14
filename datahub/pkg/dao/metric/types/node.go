package types

import (
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/datahub/pkg/metric"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	"sort"
)

// NodeMetricsDAO DAO interface of node metric data.
type NodeMetricsDAO interface {
	CreateMetrics(NodeMetricMap) error
	ListMetrics(ListNodeMetricsRequest) (NodeMetricMap, error)
}

type NodeMetricSample struct {
	NodeName   metadata.NodeName
	MetricType metric.NodeMetricType
	Metrics    []metric.Sample
}

// NodeMetric Metric model to represent one node metric
type NodeMetric struct {
	NodeName metadata.NodeName
	Metrics  map[metric.NodeMetricType][]metric.Sample
}

// NodesMetricMap Nodes' metric map
type NodeMetricMap struct {
	MetricMap map[metadata.NodeName]*NodeMetric
}

// ListNodeMetricsRequest Argument of method ListNodeMetrics
type ListNodeMetricsRequest struct {
	DBCommon.QueryCondition
	NodeNames []metadata.NodeName
}

func NewNodeMetricSample() *NodeMetricSample {
	metricSample := &NodeMetricSample{}
	metricSample.Metrics = make([]metric.Sample, 0)
	return metricSample
}

func NewNodeMetric() *NodeMetric {
	nodeMetric := &NodeMetric{}
	nodeMetric.Metrics = make(map[metric.NodeMetricType][]metric.Sample)
	return nodeMetric
}

func NewNodeMetricMap() NodeMetricMap {
	nodeMetricMap := NodeMetricMap{}
	nodeMetricMap.MetricMap = make(map[metadata.NodeName]*NodeMetric)
	return nodeMetricMap
}

// GetNodeNames Return nodes name in request
func (r ListNodeMetricsRequest) GetNodeNames() []metadata.NodeName {
	return r.NodeNames
}

// GetEmptyNodeNames Return slice with one empty string element
func (r ListNodeMetricsRequest) GetEmptyNodeNames() []metadata.NodeName {
	return []metadata.NodeName{""}
}

func (n *NodeMetric) GetSamples(metricType metric.NodeMetricType) *NodeMetricSample {
	nodeSample := NewNodeMetricSample()
	nodeSample.NodeName = n.NodeName
	nodeSample.MetricType = metricType

	if value, exist := n.Metrics[metricType]; exist {
		nodeSample.Metrics = value
	}

	return nodeSample
}

func (n *NodeMetric) AddSample(metricType metric.NodeMetricType, sample metric.Sample) {
	if _, exist := n.Metrics[metricType]; !exist {
		n.Metrics[metricType] = make([]metric.Sample, 0)
	}
	n.Metrics[metricType] = append(n.Metrics[metricType], sample)
}

// Merge Merge current NodeMetric with input NodeMetric
func (n *NodeMetric) Merge(in *NodeMetric) {
	for metricType, metrics := range in.Metrics {
		n.Metrics[metricType] = append(n.Metrics[metricType], metrics...)
	}
}

// SortByTimestamp Sort each metric samples by timestamp in input order
func (n *NodeMetric) SortByTimestamp(order DBCommon.Order) {
	for _, samples := range n.Metrics {
		if order == DBCommon.Asc {
			sort.Sort(metric.SamplesByAscTimestamp(samples))
		} else {
			sort.Sort(metric.SamplesByDescTimestamp(samples))
		}
	}
}

// Limit Slicing each metric samples element
func (n *NodeMetric) Limit(limit int) {
	if limit == 0 {
		return
	}

	for metricType, samples := range n.Metrics {
		n.Metrics[metricType] = samples[:limit]
	}
}

// AddNodeMetric Add node metric into NodesMetricMap
func (n *NodeMetricMap) AddNodeMetric(nodeMetric *NodeMetric) {
	nodeName := nodeMetric.NodeName
	if existNodeMetric, exist := n.MetricMap[nodeName]; exist {
		existNodeMetric.Merge(nodeMetric)
	} else {
		n.MetricMap[nodeName] = nodeMetric
	}
}

func (n *NodeMetricMap) GetSamples(metricType metric.NodeMetricType) []*NodeMetricSample {
	nodeSample := make([]*NodeMetricSample, 0)

	for _, nodeMetric := range n.MetricMap {
		nodeSample = append(nodeSample, nodeMetric.GetSamples(metricType))
	}

	return nodeSample
}

// SortByTimestamp Sort each node metric's content
func (n *NodeMetricMap) SortByTimestamp(order DBCommon.Order) {
	for _, nodeMetric := range n.MetricMap {
		nodeMetric.SortByTimestamp(order)
	}
}

// Limit Limit each node metric's content
func (n *NodeMetricMap) Limit(limit int) {
	for _, nodeMetric := range n.MetricMap {
		nodeMetric.Limit(limit)
	}
}
