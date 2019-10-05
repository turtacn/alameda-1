package types

import (
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/datahub/pkg/metric"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	"sort"
)

// NodeMetricsDAO DAO interface of node metric data.
type NodeMetricsDAO interface {
	ListMetrics(ListNodeMetricsRequest) (NodesMetricMap, error)
}

// ListNodeMetricsRequest Argument of method ListNodeMetrics
type ListNodeMetricsRequest struct {
	NodeNames []metadata.NodeName
	DBCommon.QueryCondition
}

// GetNodeNames Return nodes name in request
func (r ListNodeMetricsRequest) GetNodeNames() []metadata.NodeName {
	return r.NodeNames
}

// GetEmptyNodeNames Return slice with one empty string element
func (r ListNodeMetricsRequest) GetEmptyNodeNames() []metadata.NodeName {
	return []metadata.NodeName{""}
}

// NodeMetric Metric model to represent one node metric
type NodeMetric struct {
	NodeName metadata.NodeName
	Metrics  map[metric.NodeMetricType][]metric.Sample
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

// NodesMetricMap Nodes' metric map
type NodesMetricMap map[metadata.NodeName]*NodeMetric

// AddNodeMetric Add node metric into NodesMetricMap
func (n *NodesMetricMap) AddNodeMetric(nodeMetric *NodeMetric) {

	nodeName := nodeMetric.NodeName
	if existNodeMetric, exist := (*n)[nodeName]; exist {
		existNodeMetric.Merge(nodeMetric)
	} else {
		(*n)[nodeName] = nodeMetric
	}
}

// SortByTimestamp Sort each node metric's content
func (n *NodesMetricMap) SortByTimestamp(order DBCommon.Order) {

	for _, nodeMetric := range *n {
		nodeMetric.SortByTimestamp(order)
	}
}

// Limit Limit each node metric's content
func (n *NodesMetricMap) Limit(limit int) {

	for _, nodeMetric := range *n {
		nodeMetric.Limit(limit)
	}
}
