package metrics

import (
	"context"
	"fmt"
	"strings"
	"time"

	EntityPromthMetric "github.com/containers-ai/alameda/datahub/pkg/dao/entities/prometheus/metrics"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/pkg/errors"
)

// NamespaceMemoryUsageRepository Repository to access metric from prometheus
type NamespaceMemoryUsageRepository struct {
	PrometheusConfig InternalPromth.Config
}

// NewNamespaceMemoryUsageRepositoryWithConfig New namespace memory usage bytes repository with prometheus configuration
func NewNamespaceMemoryUsageRepositoryWithConfig(cfg InternalPromth.Config) NamespaceMemoryUsageRepository {
	return NamespaceMemoryUsageRepository{PrometheusConfig: cfg}
}

func (n NamespaceMemoryUsageRepository) ListNamespaceMemoryUsageBytesEntitiesByNamespaceNames(ctx context.Context, namespaceNames []string, options ...DBCommon.Option) ([]EntityPromthMetric.NamespaceMemoryUsageBytesEntity, error) {
	// Example of expression to query prometheus
	// sum(container_memory_usage_bytes{pod_name!="",container_name!="",container_name!="POD",namespace=~"@n1"}) by (namespace)

	prometheusClient, err := InternalPromth.NewClient(&n.PrometheusConfig)
	if err != nil {
		return nil, errors.Wrap(err, "new prometheus client failed")
	}

	opt := DBCommon.NewDefaultOptions()
	for _, option := range options {
		option(&opt)
	}

	queryLabelsString := n.buildQueryLabelsStringByNamespaceNames(namespaceNames)
	queryExpression := fmt.Sprintf(`%s{%s}`, ContainerMemoryUsageBytesMetricName, queryLabelsString)
	stepTimeInSeconds := int64(opt.StepTime.Nanoseconds() / int64(time.Second))
	queryExpression, err = InternalPromth.WrapQueryExpression(queryExpression, opt.AggregateOverTimeFunc, stepTimeInSeconds)
	if err != nil {
		return nil, errors.Wrap(err, "wrap query expression failed")
	}
	queryExpression = fmt.Sprintf(`sum(%s) by (%s)`, queryExpression, ContainerMemoryUsageBytesLabelNamespace)

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
	namespaceMemoryUsageBytesEntities := make([]EntityPromthMetric.NamespaceMemoryUsageBytesEntity, len(entities))
	for i, entity := range entities {
		e := n.newNamespaceMemoryUsageBytesEntity(entity)
		namespaceMemoryUsageBytesEntities[i] = e
		foundMap[e.NamespaceName] = true
	}
	for name, exist := range foundMap {
		if !exist {
			namespaceMemoryUsageBytesEntities = append(namespaceMemoryUsageBytesEntities, EntityPromthMetric.NamespaceMemoryUsageBytesEntity{
				NamespaceName: name,
			})
		}
	}

	return namespaceMemoryUsageBytesEntities, nil
}

func (n NamespaceMemoryUsageRepository) buildDefaultQueryLabelsString() string {

	var queryLabelsString = ""
	queryLabelsString += fmt.Sprintf(`%s != "",`, ContainerMemoryUsageBytesLabelPodName)
	queryLabelsString += fmt.Sprintf(`%s != "",`, ContainerMemoryUsageBytesLabelContainerName)
	queryLabelsString += fmt.Sprintf(`%s != "POD"`, ContainerMemoryUsageBytesLabelContainerName)
	return queryLabelsString
}

func (n NamespaceMemoryUsageRepository) buildQueryLabelsStringByNamespaceNames(namespaceNames []string) string {
	var (
		queryLabelsString = n.buildDefaultQueryLabelsString()
	)

	names := ""
	for _, name := range namespaceNames {
		names += fmt.Sprintf("%s|", name)
	}
	if names != "" {
		names = strings.TrimSuffix(names, "|")
		queryLabelsString += fmt.Sprintf(`,%s =~ "%s"`, ContainerMemoryUsageBytesLabelNamespace, names)
	}

	return queryLabelsString
}

func (n NamespaceMemoryUsageRepository) newNamespaceMemoryUsageBytesEntity(e InternalPromth.Entity) EntityPromthMetric.NamespaceMemoryUsageBytesEntity {

	samples := make([]FormatTypes.Sample, len(e.Values))
	for i, value := range e.Values {
		samples[i] = FormatTypes.Sample{
			Timestamp: value.UnixTime,
			Value:     value.SampleValue,
		}
	}
	return EntityPromthMetric.NamespaceMemoryUsageBytesEntity{
		PrometheusEntity: e,
		NamespaceName:    e.Labels[ContainerCpuUsagePercentageLabelNamespace],
		Samples:          samples,
	}
}
