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

// PodCPUUsageRepository Repository to access metric namespace_pod_name_container_name:container_cpu_usage_seconds_total:sum_rate from prometheus
type PodCPUUsageRepository struct {
	PrometheusConfig InternalPromth.Config
}

// NewPodCPUUsageRepositoryWithConfig New pod cpu usage millicores repository with prometheus configuration
func NewPodCPUUsageRepositoryWithConfig(cfg InternalPromth.Config) PodCPUUsageRepository {
	return PodCPUUsageRepository{PrometheusConfig: cfg}
}

func (c PodCPUUsageRepository) ListPodCPUUsageMillicoresEntitiesBySummingPodMetrics(ctx context.Context, namespace string, podNames []string, options ...DBCommon.Option) ([]EntityPromthMetric.PodCPUUsageMillicoresEntity, error) {
	// Example of expression to query prometheus
	// 1000 * sum(namespace_pod_name_container_name:container_cpu_usage_seconds_total:sum_rate{pod_name!="",container_name!="POD",namespace="@n1",pod_name=~"@p1|@p2"})

	prometheusClient, err := InternalPromth.NewClient(&c.PrometheusConfig)
	if err != nil {
		return nil, errors.Wrap(err, "new prometheus client failed")
	}

	opt := DBCommon.NewDefaultOptions()
	for _, option := range options {
		option(&opt)
	}

	queryLabelsString := c.buildDefaultQueryLabelsString()
	queryLabelsString += fmt.Sprintf(`, %s = "%s"`, ContainerCpuUsagePercentageLabelNamespace, namespace)
	names := ""
	for _, name := range podNames {
		names += fmt.Sprintf("%s|", name)
	}
	if names != "" {
		names = strings.TrimSuffix(names, "|")
		queryLabelsString += fmt.Sprintf(`,%s =~ "%s"`, ContainerCpuUsagePercentageLabelPodName, names)
	}
	queryExpression := fmt.Sprintf(`%s{%s}`, ContainerCpuUsagePercentageMetricName, queryLabelsString)
	stepTimeInSeconds := int64(opt.StepTime.Nanoseconds() / int64(time.Second))
	queryExpression, err = InternalPromth.WrapQueryExpression(queryExpression, opt.AggregateOverTimeFunc, stepTimeInSeconds)
	if err != nil {
		return nil, errors.Wrap(err, "wrap query expression failed")
	}
	queryExpression = fmt.Sprintf(`1000 * sum(%s)`, queryExpression)

	scope.Debugf("Query to prometheus: queryExpression: %+v, StartTime: %+v, EndTime: %+v, StepTime: %+v", queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
	response, err := prometheusClient.QueryRange(ctx, queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
	if err != nil {
		return nil, errors.Wrap(err, "query prometheus failed")
	} else if response.Status != InternalPromth.StatusSuccess {
		return nil, errors.Errorf("query prometheus failed: receive error response from prometheus: %s", response.Error)
	}

	entities, err := response.GetEntities()
	if err != nil {
		return nil, errors.Wrap(err, "get prometheus response entities failed")
	}
	podCPUUsageMillicoresEntities := make([]EntityPromthMetric.PodCPUUsageMillicoresEntity, len(entities))
	for i, entity := range entities {
		e := c.newPodCPUUsageMillicoresEntity(entity)
		podCPUUsageMillicoresEntities[i] = e
	}

	return podCPUUsageMillicoresEntities, nil
}

func (c PodCPUUsageRepository) buildDefaultQueryLabelsString() string {
	// 1000 * sum( {pod_name!="",container_name!="POD",namespace="@n1",pod_name=~"@p1|@p2"})

	var queryLabelsString = ""
	queryLabelsString += fmt.Sprintf(`%s != "",`, ContainerCpuUsagePercentageLabelPodName)
	queryLabelsString += fmt.Sprintf(`%s != "POD"`, ContainerCpuUsagePercentageLabelContainerName)
	return queryLabelsString
}

func (c PodCPUUsageRepository) newPodCPUUsageMillicoresEntity(e InternalPromth.Entity) EntityPromthMetric.PodCPUUsageMillicoresEntity {

	samples := make([]FormatTypes.Sample, len(e.Values))
	for i, value := range e.Values {
		samples[i] = FormatTypes.Sample{
			Timestamp: value.UnixTime,
			Value:     value.SampleValue,
		}
	}
	return EntityPromthMetric.PodCPUUsageMillicoresEntity{
		NamespaceName: e.Labels[ContainerCpuUsagePercentageLabelNamespace],
		PodName:       e.Labels[ContainerCpuUsagePercentageLabelPodName],
		Samples:       samples,
	}
}
