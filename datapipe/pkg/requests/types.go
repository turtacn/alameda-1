package requests

import (
	metric "github.com/containers-ai/alameda/datapipe/pkg/apis/metrics/define"
	metric_dao "github.com/containers-ai/alameda/datapipe/pkg/dao/metrics"
	datahubMetricsAPI "github.com/containers-ai/api/datahub/metrics"
	datahubResourceAPI "github.com/containers-ai/api/datahub/resources"
	"github.com/golang/protobuf/ptypes"
)

type DaoPodMetricExtended struct {
	*metric_dao.PodMetric
}

func (p DaoPodMetricExtended) DatahubPodMetric() *datahubMetricsAPI.PodMetric {

	var (
		datahubPodMetric datahubMetricsAPI.PodMetric
	)

	datahubPodMetric = datahubMetricsAPI.PodMetric{
		NamespacedName: &datahubResourceAPI.NamespacedName{
			Namespace: string(p.Namespace),
			Name:      string(p.PodName),
		},
	}

	for _, containerMetric := range *p.ContainersMetricMap {
		containerMetricExtended := DaoContainerMetricExtended{containerMetric}
		datahubContainerMetric := containerMetricExtended.DatahubContainerMetric()
		datahubPodMetric.ContainerMetrics = append(datahubPodMetric.ContainerMetrics, datahubContainerMetric)
	}

	return &datahubPodMetric
}

type DaoContainerMetricExtended struct {
	*metric_dao.ContainerMetric
}

func (c DaoContainerMetricExtended) DatahubContainerMetric() *datahubMetricsAPI.ContainerMetric {
	datahubContainerMetric := &datahubMetricsAPI.ContainerMetric{
		Name:       string(c.ContainerName),
		MetricData: map[int32]*datahubMetricsAPI.MetricData{},
	}

	for metricType, samples := range c.Metrics {
		if datahubMetricType, exist := metric.TypeToDatahubMetricType[metricType]; exist {
			datahubContainerMetric.MetricData[int32(datahubMetricType)] = ProduceDatahubMetricDataFromSamples(datahubMetricType, samples)
		}
	}

	return datahubContainerMetric
}

type DaoNodeMetricExtended struct {
	*metric_dao.NodeMetric
}

func (n DaoNodeMetricExtended) DatahubNodeMetric() *datahubMetricsAPI.NodeMetric {
	datahubNodeMetric := &datahubMetricsAPI.NodeMetric{
		Name:       n.NodeName,
		MetricData: map[int32]*datahubMetricsAPI.MetricData{},
	}

	for metricType, samples := range n.Metrics {
		if datahubMetricType, exist := metric.TypeToDatahubMetricType[metricType]; exist {
			datahubNodeMetric.MetricData[int32(datahubMetricType)] = ProduceDatahubMetricDataFromSamples(datahubMetricType, samples)
		}
	}

	return datahubNodeMetric
}

func ProduceDatahubMetricDataFromSamples(metricType datahubMetricsAPI.MetricType, samples []metric.Sample) *datahubMetricsAPI.MetricData {

	datahubMetricData := &datahubMetricsAPI.MetricData{}

	for _, sample := range samples {
		googleTimestamp, err := ptypes.TimestampProto(sample.Timestamp)
		if err != nil {
			googleTimestamp = nil
		}

		datahubSample := datahubMetricsAPI.Sample{StartTime: googleTimestamp, NumValue: sample.Value}
		datahubMetricData.Data = append(datahubMetricData.Data, &datahubSample)
	}

	return datahubMetricData
}
