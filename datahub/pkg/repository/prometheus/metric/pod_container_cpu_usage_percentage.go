package metric

import (
	"errors"
	"fmt"
	"time"

	"github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/containerCPUUsagePercentage"
	"github.com/containers-ai/alameda/datahub/pkg/repository/prometheus"
)

// PodContainerCPUUsagePercentageRepository comment
type PodContainerCPUUsagePercentageRepository struct {
	PrometheusConfig prometheus.Config
}

func NewPodContainerCPUUsagePercentageRepositoryWithConfig(cfg prometheus.Config) PodContainerCPUUsagePercentageRepository {
	return PodContainerCPUUsagePercentageRepository{PrometheusConfig: cfg}
}

// ListMetricsByPodNamespacedName comment
func (c PodContainerCPUUsagePercentageRepository) ListMetricsByPodNamespacedName(namespace string, podName string, startTime, endTime time.Time) ([]prometheus.Entity, error) {

	var (
		err error

		prometheusClient *prometheus.Prometheus

		metricName        string
		queryLabelsString string
		queryExpression   string

		response prometheus.Response

		entities []prometheus.Entity
	)

	prometheusClient, err = prometheus.New(c.PrometheusConfig)
	if err != nil {
		return entities, errors.New("QueryRangeByPodNamespacedName failed: " + err.Error())
	}

	metricName = containerCPUUsagePercentage.MetricName
	queryLabelsString = c.buildQueryLabelsStringByNamespaceAndPodName(namespace, podName)

	if queryLabelsString != "" {
		queryExpression = fmt.Sprintf("%s{%s}", metricName, queryLabelsString)
	} else {
		queryExpression = fmt.Sprintf("%s", metricName)
	}

	response, err = prometheusClient.QueryRange(queryExpression, startTime, endTime)
	if err != nil {
		return entities, errors.New("QueryRangeByPodNamespacedName failed: " + err.Error())
	} else if response.Status != prometheus.SuccessStatus {
		return entities, errors.New("QueryRangeByPodNamespacedName failed: receive error response from prometheus: " + response.Error)
	}

	entities, err = response.GetEntitis()
	if err != nil {
		return entities, errors.New("ListMetricsByPodNamespacedName failed: " + err.Error())
	}

	return entities, nil
}

func (c PodContainerCPUUsagePercentageRepository) buildDefaultQueryLabelsString() string {

	var queryLabelsString = ""

	queryLabelsString += fmt.Sprintf(`%s != "",`, containerCPUUsagePercentage.PodLabelName)
	queryLabelsString += fmt.Sprintf(`%s != "POD"`, containerCPUUsagePercentage.ContainerLabel)

	return queryLabelsString
}

func (c PodContainerCPUUsagePercentageRepository) buildQueryLabelsStringByNamespaceAndPodName(namespace string, podName string) string {

	var (
		queryLabelsString = c.buildDefaultQueryLabelsString()
	)

	if namespace != "" {
		queryLabelsString += fmt.Sprintf(`,%s = "%s"`, containerCPUUsagePercentage.NamespaceLabel, namespace)
	}

	if podName != "" {
		queryLabelsString += fmt.Sprintf(`,%s = "%s"`, containerCPUUsagePercentage.PodLabelName, podName)
	}

	return queryLabelsString
}
