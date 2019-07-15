package influxdb

import (
	RepoInfluxMetric "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb/metric"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

type influxdbMetricDAOImpl struct {
	influxDBConfig InternalInflux.Config
}

func NewWithConfig(config InternalInflux.Config) *influxdbMetricDAOImpl {
	return &influxdbMetricDAOImpl{influxDBConfig: config}
}

func (i influxdbMetricDAOImpl) ListContainerMetrics(in *datahub_v1alpha1.ListPodMetricsRequest) ([]*datahub_v1alpha1.PodMetric, error) {
	metricRepo := RepoInfluxMetric.NewContainerRepositoryWithConfig(i.influxDBConfig)
	return metricRepo.ListContainerMetrics(in)
}

func (i influxdbMetricDAOImpl) ListNodeMetrics(in *datahub_v1alpha1.ListNodeMetricsRequest) ([]*datahub_v1alpha1.NodeMetric, error) {
	metricRepo := RepoInfluxMetric.NewNodeRepositoryWithConfig(i.influxDBConfig)
	return metricRepo.ListNodeMetrics(in)
}
