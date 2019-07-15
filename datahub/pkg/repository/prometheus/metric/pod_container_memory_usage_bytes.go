package metric

import (
	"fmt"
	EntityPromthContainerMemUsage "github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/containerMemoryUsageBytes"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/pkg/errors"
	"time"
)

// PodContainerMemoryUsageBytesRepository Repository to access metric container_memory_usage_bytes from prometheus
type PodContainerMemoryUsageBytesRepository struct {
	PrometheusConfig InternalPromth.Config
}

// NewPodContainerMemoryUsageBytesRepositoryWithConfig New pod container memory usage bytes repository with prometheus configuration
func NewPodContainerMemoryUsageBytesRepositoryWithConfig(cfg InternalPromth.Config) PodContainerMemoryUsageBytesRepository {
	return PodContainerMemoryUsageBytesRepository{PrometheusConfig: cfg}
}

// ListMetricsByPodNamespacedName Provide metrics from response of querying request contain namespace, pod_name and default labels
func (c PodContainerMemoryUsageBytesRepository) ListMetricsByPodNamespacedName(namespace string, podName string, options ...DBCommon.Option) ([]InternalPromth.Entity, error) {

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
		return entities, errors.Wrap(err, "list pod container memory usage metrics failed")
	}

	opt := DBCommon.NewDefaultOptions()
	for _, option := range options {
		option(&opt)
	}

	metricName = EntityPromthContainerMemUsage.MetricName
	queryLabelsString = c.buildQueryLabelsStringByNamespaceAndPodName(namespace, podName)

	if queryLabelsString != "" {
		queryExpression = fmt.Sprintf("%s{%s}", metricName, queryLabelsString)
	} else {
		queryExpression = fmt.Sprintf("%s", metricName)
	}

	stepTimeInSeconds := int64(opt.StepTime.Nanoseconds() / int64(time.Second))
	queryExpression, err = InternalPromth.WrapQueryExpression(queryExpression, opt.AggregateOverTimeFunc, stepTimeInSeconds)
	if err != nil {
		return entities, errors.Wrap(err, "list pod container memory usage metric by namespaced name failed")
	}

	response, err = prometheusClient.QueryRange(queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
	if err != nil {
		return entities, errors.Wrap(err, "list pod container memory usage metrics failed")
	} else if response.Status != InternalPromth.StatusSuccess {
		return entities, errors.Errorf("list pod container memory usage metrics failed: receive error response from prometheus: %s", response.Error)
	}

	entities, err = response.GetEntities()
	if err != nil {
		return entities, errors.Wrap(err, "list pod container memory usage metrics failed")
	}

	return entities, nil
}

func (c PodContainerMemoryUsageBytesRepository) buildDefaultQueryLabelsString() string {

	var queryLabelsString = ""

	queryLabelsString += fmt.Sprintf(`%s != "" ,`, EntityPromthContainerMemUsage.PodLabelName)
	queryLabelsString += fmt.Sprintf(`%s != "" ,`, EntityPromthContainerMemUsage.ContainerLabel)
	queryLabelsString += fmt.Sprintf(`%s != "POD"`, EntityPromthContainerMemUsage.ContainerLabel)

	return queryLabelsString
}

func (c PodContainerMemoryUsageBytesRepository) buildQueryLabelsStringByNamespaceAndPodName(namespace string, podName string) string {

	var (
		queryLabelsString = c.buildDefaultQueryLabelsString()
	)

	if namespace != "" {
		queryLabelsString += fmt.Sprintf(`,%s = "%s"`, EntityPromthContainerMemUsage.NamespaceLabel, namespace)
	}

	if podName != "" {
		queryLabelsString += fmt.Sprintf(`,%s = "%s"`, EntityPromthContainerMemUsage.PodLabelName, podName)
	}

	return queryLabelsString
}
