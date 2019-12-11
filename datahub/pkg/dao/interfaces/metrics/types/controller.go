package types

import (
	"context"
	"sort"

	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
)

// ControllerMetricsDAO DAO interface of app metric data.
type ControllerMetricsDAO interface {
	CreateMetrics(context.Context, ControllerMetricMap) error
	ListMetrics(context.Context, ListControllerMetricsRequest) (ControllerMetricMap, error)
}

type ControllerMetricSample struct {
	ObjectMeta ControllerObjectMeta
	MetricType enumconv.MetricType
	Metrics    []types.Sample
}

type ControllerObjectMeta struct {
	metadata.ObjectMeta
	Kind string
}

// ListControllerMetricsRequest Argument of method ListMetrics
// ControllerObjectMetas is used to assign to list specific apps metrics
type ListControllerMetricsRequest struct {
	common.QueryCondition
	ObjectMetas []metadata.ObjectMeta
	Kind        string // DEPLOYMENT, DEPLOYMENTCONFIG and STATEFULSET
}

// ControllerMetric Metric model to represent one app metric
type ControllerMetric struct {
	ObjectMeta ControllerObjectMeta
	Metrics    map[enumconv.MetricType][]types.Sample
}

func NewControllerMetric() *ControllerMetric {
	metric := &ControllerMetric{}
	metric.Metrics = make(map[enumconv.MetricType][]types.Sample)
	return metric
}

func (c *ControllerMetric) AddSample(metricType enumconv.MetricType, sample types.Sample) {
	if c.Metrics == nil {
		c.Metrics = make(map[enumconv.MetricType][]types.Sample)
	}
	if _, exist := c.Metrics[metricType]; !exist {
		c.Metrics[metricType] = make([]types.Sample, 0)
	}
	c.Metrics[metricType] = append(c.Metrics[metricType], sample)
}

func (c *ControllerMetric) GetSamples(metricType enumconv.MetricType) ControllerMetricSample {
	if c.Metrics == nil {
		c.Metrics = make(map[enumconv.MetricType][]types.Sample)
	}
	return ControllerMetricSample{
		ObjectMeta: c.ObjectMeta,
		MetricType: metricType,
		Metrics:    c.Metrics[metricType],
	}
}

// Merge Merge current ControllerMetricMap with input ControllerMetricMap
func (c *ControllerMetric) Merge(in *ControllerMetric) {
	if c.Metrics == nil {
		c.Metrics = make(map[enumconv.MetricType][]types.Sample)
	}
	for metricType, metric := range in.Metrics {
		if _, exist := c.Metrics[metricType]; exist {
			c.Metrics[metricType] = append(c.Metrics[metricType], metric...)
		} else {
			c.Metrics[metricType] = metric
		}
	}
}

// SortByTimestamp Sort each metric samples by timestamp in input order
func (c *ControllerMetric) SortByTimestamp(order common.Order) {
	for _, samples := range c.Metrics {
		if order == common.Asc {
			sort.Sort(types.SamplesByAscTimestamp(samples))
		} else {
			sort.Sort(types.SamplesByDescTimestamp(samples))
		}
	}
}

// Limit Slicing each metric samples element
func (c *ControllerMetric) Limit(limit int) {
	if limit == 0 {
		return
	}

	for metricType, samples := range c.Metrics {
		if len(samples) < limit {
			continue
		}
		c.Metrics[metricType] = samples[:limit]
	}
}

// ControllerMetricMap Controllers metric map
type ControllerMetricMap struct {
	MetricMap map[ControllerObjectMeta]*ControllerMetric
}

func NewControllerMetricMap() ControllerMetricMap {
	metricMap := ControllerMetricMap{}
	metricMap.MetricMap = make(map[ControllerObjectMeta]*ControllerMetric)
	return metricMap
}

func (c *ControllerMetricMap) AddControllerMetric(metric *ControllerMetric) {
	if c.MetricMap == nil {
		c.MetricMap = make(map[ControllerObjectMeta]*ControllerMetric)
	}
	if metricMap, exist := c.MetricMap[metric.ObjectMeta]; exist {
		metricMap.Merge(metric)
	} else {
		c.MetricMap[metric.ObjectMeta] = metric
	}
}

func (c *ControllerMetricMap) GetSamples(metricType enumconv.MetricType) []ControllerMetricSample {
	controllerMetricSamples := make([]ControllerMetricSample, 0, len(c.MetricMap))
	for _, metric := range c.MetricMap {
		if metric == nil {
			continue
		}
		controllerMetricSamples = append(controllerMetricSamples, metric.GetSamples(metricType))
	}

	return controllerMetricSamples
}

// SortByTimestamp Sort each node metric's content
func (c *ControllerMetricMap) SortByTimestamp(order common.Order) {
	for _, m := range c.MetricMap {
		m.SortByTimestamp(order)
	}
}

// Limit Limit each node metric's content
func (c *ControllerMetricMap) Limit(limit int) {
	for _, m := range c.MetricMap {
		m.Limit(limit)
	}
}
