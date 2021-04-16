package metric

import (
	"fmt"
	EntityPromthContainerCpuUsage "github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/containerCPUUsagePercentage"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/containers-ai/alameda/pkg/utils/log"
	"github.com/pkg/errors"
	"time"
)

var (
	pod_container_memory_usage_percentage_scope = log.RegisterScope("", "", 0)
)

// PodContainerCPUUsagePercentageRepository Repository to access metric namespace_pod_name_container_name:container_cpu_usage_seconds_total:sum_rate from prometheus
type PodContainerCPUUsagePercentageRepository struct {
	PrometheusConfig InternalPromth.Config
}

// NewPodContainerCPUUsagePercentageRepositoryWithConfig New pod container cpu usage percentage repository with prometheus configuration
func NewPodContainerCPUUsagePercentageRepositoryWithConfig(cfg InternalPromth.Config) PodContainerCPUUsagePercentageRepository {
	return PodContainerCPUUsagePercentageRepository{PrometheusConfig: cfg}
}

// ListMetricsByPodNamespacedName Provide metrics from response of querying request contain namespace, pod_name and default labels
func (c PodContainerCPUUsagePercentageRepository) ListMetricsByPodNamespacedName(namespace string, podName string, options ...DBCommon.Option) ([]InternalPromth.Entity, error) {

	pod_container_memory_usage_percentage_scope.Infof("pod_container_memory_usage_percentage_scope metric-ListMetricsByPodNamespacedName input ns %s; podname %s", namespace, podName)
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
		pod_container_memory_usage_percentage_scope.Errorf("pod_container_memory_usage_percentage_scope metric-ListMetricsByPodNamespacedName error %v", err)
		return entities, errors.Wrap(err, "list pod container cpu usage metric by namespaced name failed")
	}

	opt := DBCommon.NewDefaultOptions()
	for _, option := range options {
		option(&opt)
	}

	metricName = EntityPromthContainerCpuUsage.MetricName
	queryLabelsString = c.buildQueryLabelsStringByNamespaceAndPodName(namespace, podName)

	if queryLabelsString != "" {
		queryExpression = fmt.Sprintf("%s{%s}", metricName, queryLabelsString)
	} else {
		queryExpression = fmt.Sprintf("%s", metricName)
	}

	stepTimeInSeconds := int64(opt.StepTime.Nanoseconds() / int64(time.Second))
	queryExpression, err = InternalPromth.WrapQueryExpression(queryExpression, opt.AggregateOverTimeFunc, stepTimeInSeconds)
	if err != nil {
		pod_container_memory_usage_percentage_scope.Errorf("pod_container_memory_usage_percentage_scope metric-ListMetricsByPodNamespacedName error %v", err)
		return entities, errors.Wrap(err, "list pod container cpu usage metric by namespaced name failed")
	}

	if opt.StartTime == nil {
		newS := time.Now().Add(time.Duration(-3600) * time.Second)
		opt.StartTime = &newS
	}
	response, err = prometheusClient.QueryRange(queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
	if err != nil {
		pod_container_memory_usage_percentage_scope.Errorf("pod_container_memory_usage_percentage_scope metric-ListMetricsByPodNamespacedName error %v", err)
		return entities, errors.Wrap(err, "list pod container cpu usage metric by namespaced name failed")
	} else if response.Status != InternalPromth.StatusSuccess {
		// 业务不成功
		pod_container_memory_usage_percentage_scope.Errorf("pod_container_memory_usage_percentage_scope metric-ListMetricsByPodNamespacedName not success, response [%s]", response.Error)
		return entities, errors.Errorf("list pod container cpu usage metric by namespaced name failed: receive error response from prometheus: %s", response.Error)
	}

	entities, err = response.GetEntities()
	if err != nil {
		pod_container_memory_usage_percentage_scope.Errorf("pod_container_memory_usage_percentage_scope metric-ListMetricsByPodNamespacedName error %v", err)
		return entities, errors.Wrap(err, "list pod container cpu usage metric by namespaced name failed")
	}

	pod_container_memory_usage_percentage_scope.Infof("pod_container_memory_usage_percentage_scope metric-ListMetricsByPodNamespacedName return %d %v", len(entities), &entities)
	return entities, nil
}

func (c PodContainerCPUUsagePercentageRepository) buildDefaultQueryLabelsString() string {

	var queryLabelsString = ""

	queryLabelsString += fmt.Sprintf(`%s != "",`, EntityPromthContainerCpuUsage.PodLabelName)
	queryLabelsString += fmt.Sprintf(`%s != "POD"`, EntityPromthContainerCpuUsage.ContainerLabel)

	return queryLabelsString
}

func (c PodContainerCPUUsagePercentageRepository) buildQueryLabelsStringByNamespaceAndPodName(namespace string, podName string) string {

	var (
		queryLabelsString = c.buildDefaultQueryLabelsString()
	)

	if namespace != "" {
		queryLabelsString += fmt.Sprintf(`,%s = "%s"`, EntityPromthContainerCpuUsage.NamespaceLabel, namespace)
	}

	if podName != "" {
		queryLabelsString += fmt.Sprintf(`,%s = "%s"`, EntityPromthContainerCpuUsage.PodLabelName, podName)
	}

	return queryLabelsString
}
