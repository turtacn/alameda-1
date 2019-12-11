package types

import (
	"context"
	"sort"

	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/enumconv"
	"github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	"github.com/containers-ai/alameda/internal/pkg/database/common"
)

// AppMetricsDAO DAO interface of app metric data.
type AppMetricsDAO interface {
	CreateMetrics(context.Context, AppMetricMap) error
	ListMetrics(context.Context, ListAppMetricsRequest) (AppMetricMap, error)
}

// ListAppMetricsRequest Argument of method ListMetrics
// AppObjectMetas is used to assign to list specific apps metrics
type ListAppMetricsRequest struct {
	common.QueryCondition
	ObjectMetas []metadata.ObjectMeta
}

type AppMetricSample struct {
	ObjectMeta metadata.ObjectMeta
	MetricType enumconv.MetricType
	Metrics    []types.Sample
}

// AppMetric Metric model to represent one app metric
type AppMetric struct {
	ObjectMeta metadata.ObjectMeta
	Metrics    map[enumconv.MetricType][]types.Sample
}

func NewAppMetric() *AppMetric {
	metric := &AppMetric{}
	metric.Metrics = make(map[enumconv.MetricType][]types.Sample)
	return metric
}

func (c *AppMetric) AddSample(metricType enumconv.MetricType, sample types.Sample) {
	if c.Metrics == nil {
		c.Metrics = make(map[enumconv.MetricType][]types.Sample)
	}
	if _, exist := c.Metrics[metricType]; !exist {
		c.Metrics[metricType] = make([]types.Sample, 0)
	}
	c.Metrics[metricType] = append(c.Metrics[metricType], sample)
}

// Merge Merge current AppMetricMap with input AppMetricMap
func (c *AppMetric) Merge(in *AppMetric) {
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

func (c *AppMetric) GetSamples(metricType enumconv.MetricType) AppMetricSample {
	if c.Metrics == nil {
		c.Metrics = make(map[enumconv.MetricType][]types.Sample)
	}
	return AppMetricSample{
		ObjectMeta: c.ObjectMeta,
		MetricType: metricType,
		Metrics:    c.Metrics[metricType],
	}
}

// SortByTimestamp Sort each metric samples by timestamp in input order
func (c *AppMetric) SortByTimestamp(order common.Order) {
	for _, samples := range c.Metrics {
		if order == common.Asc {
			sort.Sort(types.SamplesByAscTimestamp(samples))
		} else {
			sort.Sort(types.SamplesByDescTimestamp(samples))
		}
	}
}

// Limit Slicing each metric samples element
func (c *AppMetric) Limit(limit int) {
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

// AppMetricMap Apps metric map
type AppMetricMap struct {
	MetricMap map[metadata.ObjectMeta]*AppMetric
}

func NewAppMetricMap() AppMetricMap {
	metricMap := AppMetricMap{}
	metricMap.MetricMap = make(map[metadata.ObjectMeta]*AppMetric)
	return metricMap
}

func (c *AppMetricMap) AddAppMetric(metric *AppMetric) {
	if c.MetricMap == nil {
		c.MetricMap = make(map[metadata.ObjectMeta]*AppMetric)
	}
	if existAppMetric, exist := c.MetricMap[metric.ObjectMeta]; exist {
		existAppMetric.Merge(metric)
	} else {
		c.MetricMap[metric.ObjectMeta] = metric
	}
}

func (c *AppMetricMap) GetSamples(metricType enumconv.MetricType) []AppMetricSample {
	appMetricSamples := make([]AppMetricSample, 0, len(c.MetricMap))
	for _, metric := range c.MetricMap {
		if metric == nil {
			continue
		}
		appMetricSamples = append(appMetricSamples, metric.GetSamples(metricType))
	}

	return appMetricSamples
}

// SortByTimestamp Sort each node metric's content
func (c *AppMetricMap) SortByTimestamp(order common.Order) {
	for _, m := range c.MetricMap {
		m.SortByTimestamp(order)
	}
}

// Limit Limit each node metric's content
func (c *AppMetricMap) Limit(limit int) {
	for _, m := range c.MetricMap {
		m.Limit(limit)
	}
}
