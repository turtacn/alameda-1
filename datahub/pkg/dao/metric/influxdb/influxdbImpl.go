package influxdb

import (
	influxdb_repository "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	influxdb_repository_metric "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb/metric"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

type influxdbMetricDAOImpl struct {
	influxDBConfig influxdb_repository.Config
}

func NewWithConfig(config influxdb_repository.Config) *influxdbMetricDAOImpl {
	return &influxdbMetricDAOImpl{influxDBConfig: config}
}

func (i influxdbMetricDAOImpl) ListContainerMetrics(in *datahub_v1alpha1.ListPodMetricsRequest) ([]*datahub_v1alpha1.PodMetric, error) {
	metricRepo := influxdb_repository_metric.NewContainerRepositoryWithConfig(i.influxDBConfig)
	return metricRepo.ListContainerMetrics(in)
}

func (i influxdbMetricDAOImpl) ListNodeMetrics(in *datahub_v1alpha1.ListNodeMetricsRequest) ([]*datahub_v1alpha1.NodeMetric, error) {
	metricRepo := influxdb_repository_metric.NewNodeRepositoryWithConfig(i.influxDBConfig)
	return metricRepo.ListNodeMetrics(in)
}
