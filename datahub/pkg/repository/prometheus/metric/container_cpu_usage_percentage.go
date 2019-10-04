package metric

import (
	"fmt"
	EntityPromthMetric "github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/metric"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/pkg/errors"
	"time"
)

// PodContainerCPUUsagePercentageRepository Repository to access metric namespace_pod_name_container_name:container_cpu_usage_seconds_total:sum_rate from prometheus
type PodContainerCpuUsagePercentageRepository struct {
	PrometheusConfig InternalPromth.Config
}

// NewPodContainerCPUUsagePercentageRepositoryWithConfig New pod container cpu usage percentage repository with prometheus configuration
func NewPodContainerCpuUsagePercentageRepositoryWithConfig(cfg InternalPromth.Config) PodContainerCpuUsagePercentageRepository {
	return PodContainerCpuUsagePercentageRepository{PrometheusConfig: cfg}
}

// ListMetricsByPodNamespacedName Provide metrics from response of querying request contain namespace, pod_name and default labels
func (c PodContainerCpuUsagePercentageRepository) ListMetricsByPodNamespacedName(namespace string, podName string, options ...DBCommon.Option) ([]InternalPromth.Entity, error) {

	var (
		err error

		prometheusClient *InternalPromth.Prometheus

		metricName        string
		queryLabelsString string
		queryExpression   string

		response InternalPromth.Response

		entities []InternalPromth.Entity
	)

	prometheusClient, err = InternalPromth.NewClient(&c.PrometheusConfig)
	if err != nil {
		return entities, errors.Wrap(err, "list pod container cpu usage metric by namespaced name failed")
	}

	opt := DBCommon.NewDefaultOptions()
	for _, option := range options {
		option(&opt)
	}

	metricName = EntityPromthMetric.ContainerCpuUsagePercentageMetricName
	queryLabelsString = c.buildQueryLabelsStringByNamespaceAndPodName(namespace, podName)

	if queryLabelsString != "" {
		queryExpression = fmt.Sprintf("%s{%s}", metricName, queryLabelsString)
	} else {
		queryExpression = fmt.Sprintf("%s", metricName)
	}

	stepTimeInSeconds := int64(opt.StepTime.Nanoseconds() / int64(time.Second))
	queryExpression, err = InternalPromth.WrapQueryExpression(queryExpression, opt.AggregateOverTimeFunc, stepTimeInSeconds)
	if err != nil {
		return entities, errors.Wrap(err, "list pod container cpu usage metric by namespaced name failed")
	}

	response, err = prometheusClient.QueryRange(queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
	if err != nil {
		return entities, errors.Wrap(err, "list pod container cpu usage metric by namespaced name failed")
	} else if response.Status != InternalPromth.StatusSuccess {
		return entities, errors.Errorf("list pod container cpu usage metric by namespaced name failed: receive error response from prometheus: %s", response.Error)
	}

	entities, err = response.GetEntities()
	if err != nil {
		return entities, errors.Wrap(err, "list pod container cpu usage metric by namespaced name failed")
	}

	return entities, nil
}

func (c PodContainerCpuUsagePercentageRepository) buildDefaultQueryLabelsString() string {

	var queryLabelsString = ""

	queryLabelsString += fmt.Sprintf(`%s != "",`, EntityPromthMetric.ContainerCpuUsagePercentageLabelPodName)
	queryLabelsString += fmt.Sprintf(`%s != "POD"`, EntityPromthMetric.ContainerCpuUsagePercentageLabelContainerName)

	return queryLabelsString
}

func (c PodContainerCpuUsagePercentageRepository) buildQueryLabelsStringByNamespaceAndPodName(namespace string, podName string) string {

	var (
		queryLabelsString = c.buildDefaultQueryLabelsString()
	)

	if namespace != "" {
		queryLabelsString += fmt.Sprintf(`,%s = "%s"`, EntityPromthMetric.ContainerCpuUsagePercentageLabelNamespace, namespace)
	}

	if podName != "" {
		queryLabelsString += fmt.Sprintf(`,%s = "%s"`, EntityPromthMetric.ContainerCpuUsagePercentageLabelPodName, podName)
	}

	return queryLabelsString
}
