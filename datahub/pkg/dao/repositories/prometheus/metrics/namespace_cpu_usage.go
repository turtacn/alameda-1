package metrics

import (
	"context"
	"fmt"
	EntityPromthMetric "github.com/containers-ai/alameda/datahub/pkg/dao/entities/prometheus/metrics"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/pkg/errors"
	"strings"
	"time"
)

// NamespaceCPUUsageRepository Repository to access metric namespace_pod_name_container_name:container_cpu_usage_seconds_total:sum_rate from prometheus
type NamespaceCPUUsageRepository struct {
	PrometheusConfig InternalPromth.Config
}

// NewNamespaceCPUUsageRepositoryWithConfig New namespace cpu usage millicores repository with prometheus configuration
func NewNamespaceCPUUsageRepositoryWithConfig(cfg InternalPromth.Config) NamespaceCPUUsageRepository {
	return NamespaceCPUUsageRepository{PrometheusConfig: cfg}
}

func (c NamespaceCPUUsageRepository) ListNamespaceCPUUsageMillicoresEntitiesByNamespaceNames(ctx context.Context, namespaceNames []string, options ...DBCommon.Option) ([]EntityPromthMetric.NamespaceCPUUsageMillicoresEntity, error) {
	// Example of expression to query prometheus
	// 1000 * sum(namespace_pod_name_container_name:container_cpu_usage_seconds_total:sum_rate{pod_name!="",container_name!="POD",namespace=~"@n1"}) by (namespace)

	prometheusClient, err := InternalPromth.NewClient(&c.PrometheusConfig)
	if err != nil {
		return nil, errors.Wrap(err, "new prometheus client failed")
	}

	opt := DBCommon.NewDefaultOptions()
	for _, option := range options {
		option(&opt)
	}

	queryLabelsString := c.buildQueryLabelsStringByNamespaceNames(namespaceNames)
	queryExpression := fmt.Sprintf("%s{%s}", ContainerCpuUsagePercentageMetricName, queryLabelsString)
	stepTimeInSeconds := int64(opt.StepTime.Nanoseconds() / int64(time.Second))
	queryExpression, err = InternalPromth.WrapQueryExpression(queryExpression, opt.AggregateOverTimeFunc, stepTimeInSeconds)
	if err != nil {
		return nil, errors.Wrap(err, "wrap query expression failed")
	}
	queryExpression = fmt.Sprintf(`1000 * sum(%s) by (%s)`, queryExpression, ContainerCpuUsagePercentageLabelNamespace)

	scope.Debugf("Query to prometheus: queryExpression: %+v, StartTime: %+v, EndTime: %+v, StepTime: %+v", queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
	response, err := prometheusClient.QueryRange(ctx, queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
	if err != nil {
		return nil, errors.Wrap(err, "query prometheus failed")
	} else if response.Status != InternalPromth.StatusSuccess {
		return nil, errors.Errorf("query prometheus failed: receive error response from prometheus: %s", response.Error)
	}

	entities, err := response.GetEntities()
	if err != nil {
		return nil, errors.Wrap(err, "get prometheus response entites failed")
	}
	foundMap := map[string]bool{}
	for _, name := range namespaceNames {
		foundMap[name] = false
	}
	namespaceCPUUsageMillicoresEntities := make([]EntityPromthMetric.NamespaceCPUUsageMillicoresEntity, len(entities))
	for i, entity := range entities {
		e := c.newNamespaceCPUUsageMillicoresEntity(entity)
		namespaceCPUUsageMillicoresEntities[i] = e
		foundMap[e.NamespaceName] = true
	}
	for name, exist := range foundMap {
		if !exist {
			namespaceCPUUsageMillicoresEntities = append(namespaceCPUUsageMillicoresEntities, EntityPromthMetric.NamespaceCPUUsageMillicoresEntity{
				NamespaceName: name,
			})
		}
	}

	return namespaceCPUUsageMillicoresEntities, nil
}

func (c NamespaceCPUUsageRepository) buildDefaultQueryLabelsString() string {
	// sum(namespace_pod_name_container_name:container_cpu_usage_seconds_total:sum_rate{pod_name!="",container_name!="POD",namespace="@n1"})

	var queryLabelsString = ""

	queryLabelsString += fmt.Sprintf(`%s != "",`, ContainerCpuUsagePercentageLabelPodName)
	queryLabelsString += fmt.Sprintf(`%s != "POD"`, ContainerCpuUsagePercentageLabelContainerName)

	return queryLabelsString
}

func (c NamespaceCPUUsageRepository) buildQueryLabelsStringByNamespaceNames(namespaceNames []string) string {
	var (
		queryLabelsString = c.buildDefaultQueryLabelsString()
	)

	names := ""
	for _, name := range namespaceNames {
		names += fmt.Sprintf("%s|", name)
	}
	if names != "" {
		names = strings.TrimSuffix(names, "|")
		queryLabelsString += fmt.Sprintf(`,%s =~ "%s"`, ContainerCpuUsagePercentageLabelNamespace, names)
	}

	return queryLabelsString
}

func (c NamespaceCPUUsageRepository) newNamespaceCPUUsageMillicoresEntity(e InternalPromth.Entity) EntityPromthMetric.NamespaceCPUUsageMillicoresEntity {

	samples := make([]FormatTypes.Sample, len(e.Values))
	for i, value := range e.Values {
		samples[i] = FormatTypes.Sample{
			Timestamp: value.UnixTime,
			Value:     value.SampleValue,
		}
	}
	return EntityPromthMetric.NamespaceCPUUsageMillicoresEntity{
		PrometheusEntity: e,
		NamespaceName:    e.Labels[ContainerCpuUsagePercentageLabelNamespace],
		Samples:          samples,
	}
}
