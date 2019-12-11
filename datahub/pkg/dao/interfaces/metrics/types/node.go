package types

import (
	"context"
	"sort"

	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
)

// NodeMetricsDAO DAO interface of node metric data.
type NodeMetricsDAO interface {
	CreateMetrics(context.Context, NodeMetricMap) error
	ListMetrics(context.Context, ListNodeMetricsRequest) (NodeMetricMap, error)
}

type NodeMetricSample struct {
	ObjectMeta metadata.ObjectMeta
	MetricType enumconv.MetricType
	Metrics    []types.Sample
}

// NodeMetric Metric model to represent one node metric
type NodeMetric struct {
	ObjectMeta metadata.ObjectMeta
	Metrics    map[enumconv.MetricType][]types.Sample
}

// NodesMetricMap Nodes' metric map
type NodeMetricMap struct {
	MetricMap map[metadata.ObjectMeta]*NodeMetric
}

// ListNodeMetricsRequest Argument of method ListNodeMetrics
type ListNodeMetricsRequest struct {
	common.QueryCondition
	ObjectMetas []metadata.ObjectMeta
}

func NewNodeMetricSample() *NodeMetricSample {
	metricSample := &NodeMetricSample{}
	metricSample.Metrics = make([]types.Sample, 0)
	return metricSample
}

func NewNodeMetric() *NodeMetric {
	nodeMetric := &NodeMetric{}
	nodeMetric.Metrics = make(map[enumconv.MetricType][]types.Sample)
	return nodeMetric
}

func NewNodeMetricMap() NodeMetricMap {
	nodeMetricMap := NodeMetricMap{}
	nodeMetricMap.MetricMap = make(map[metadata.ObjectMeta]*NodeMetric)
	return nodeMetricMap
}

func NewListNodeMetricsRequest() ListNodeMetricsRequest {
	request := ListNodeMetricsRequest{}
	request.ObjectMetas = make([]metadata.ObjectMeta, 0)
	return request
}

// GetNodeNames Return nodes name in request
func (r ListNodeMetricsRequest) GetNodeNames() []metadata.NodeName {
	nodeNames := make([]metadata.NodeName, 0)
	for _, objectMeta := range r.ObjectMetas {
		nodeNames = append(nodeNames, objectMeta.Name)
	}
	return nodeNames
}

// GetEmptyNodeNames Return slice with one empty string element
func (r ListNodeMetricsRequest) GetEmptyNodeNames() []metadata.NodeName {
	return []metadata.NodeName{""}
}

func (n *NodeMetric) GetSamples(metricType enumconv.MetricType) *NodeMetricSample {
	nodeSample := NewNodeMetricSample()
	nodeSample.ObjectMeta.Name = n.ObjectMeta.Name
	nodeSample.ObjectMeta.ClusterName = n.ObjectMeta.ClusterName
	nodeSample.MetricType = metricType

	if n.Metrics == nil {
		n.Metrics = make(map[enumconv.MetricType][]types.Sample)
	}
	if value, exist := n.Metrics[metricType]; exist {
		nodeSample.Metrics = value
	}

	return nodeSample
}

func (n *NodeMetric) AddSample(metricType enumconv.MetricType, sample types.Sample) {
	if n.Metrics == nil {
		n.Metrics = make(map[enumconv.MetricType][]types.Sample)
	}
	if _, exist := n.Metrics[metricType]; !exist {
		n.Metrics[metricType] = make([]types.Sample, 0)
	}
	n.Metrics[metricType] = append(n.Metrics[metricType], sample)
}

// Merge Merge current NodeMetric with input NodeMetric
func (n *NodeMetric) Merge(in *NodeMetric) {
	if n.Metrics == nil {
		n.Metrics = make(map[enumconv.MetricType][]types.Sample)
	}
	for metricType, metrics := range in.Metrics {
		n.Metrics[metricType] = append(n.Metrics[metricType], metrics...)
	}
}

// SortByTimestamp Sort each metric samples by timestamp in input order
func (n *NodeMetric) SortByTimestamp(order common.Order) {
	for _, samples := range n.Metrics {
		if order == common.Asc {
			sort.Sort(types.SamplesByAscTimestamp(samples))
		} else {
			sort.Sort(types.SamplesByDescTimestamp(samples))
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
	if n.MetricMap == nil {
		n.MetricMap = make(map[metadata.ObjectMeta]*NodeMetric)
	}
	if existNodeMetric, exist := n.MetricMap[nodeMetric.ObjectMeta]; exist {
		existNodeMetric.Merge(nodeMetric)
	} else {
		n.MetricMap[nodeMetric.ObjectMeta] = nodeMetric
	}
}

func (n *NodeMetricMap) GetSamples(metricType enumconv.MetricType) []*NodeMetricSample {
	nodeSample := make([]*NodeMetricSample, 0)

	for _, nodeMetric := range n.MetricMap {
		nodeSample = append(nodeSample, nodeMetric.GetSamples(metricType))
	}

	return nodeSample
}

// SortByTimestamp Sort each node metric's content
func (n *NodeMetricMap) SortByTimestamp(order common.Order) {
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
