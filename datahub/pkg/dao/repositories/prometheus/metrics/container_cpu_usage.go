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

type ContainerCpuUsageRepository struct {
	PrometheusConfig InternalPromth.Config
}

func NewContainerCpuUsageRepositoryWithConfig(cfg InternalPromth.Config) ContainerCpuUsageRepository {
	return ContainerCpuUsageRepository{PrometheusConfig: cfg}
}

func (c ContainerCpuUsageRepository) ListContainerCPUUsageMillicoresEntitiesByNamespaceAndPodNames(ctx context.Context, namespace string, podNames []string, options ...DBCommon.Option) ([]EntityPromthMetric.ContainerCPUUsageMillicoresEntity, error) {

	prometheusClient, err := InternalPromth.NewClient(&c.PrometheusConfig)
	if err != nil {
		return nil, errors.Wrap(err, "new prometheus client failed")
	}

	opt := DBCommon.NewDefaultOptions()
	for _, option := range options {
		option(&opt)
	}

	queryLabelsString := ""
	queryLabelsString += fmt.Sprintf(`%s != "",`, ContainerCpuUsagePercentageLabelPodName)
	queryLabelsString += fmt.Sprintf(`%s != "POD",`, ContainerCpuUsagePercentageLabelContainerName)
	queryLabelsString += fmt.Sprintf(`%s = "%s",`, ContainerCpuUsagePercentageLabelNamespace, namespace)
	names := ""
	for _, podName := range podNames {
		names += fmt.Sprintf("%s|", podName)
	}
	if names != "" {
		names = strings.TrimSuffix(names, "|")
		queryLabelsString += fmt.Sprintf(`%s =~ "%s",`, ContainerCpuUsagePercentageLabelPodName, names)
	}

	queryLabelsString = strings.TrimSuffix(queryLabelsString, ",")
	queryExpression := fmt.Sprintf("%s{%s}", ContainerCpuUsagePercentageMetricName, queryLabelsString)
	stepTimeInSeconds := int64(opt.StepTime.Nanoseconds() / int64(time.Second))
	queryExpression, err = InternalPromth.WrapQueryExpression(queryExpression, opt.AggregateOverTimeFunc, stepTimeInSeconds)
	if err != nil {
		return nil, errors.Wrap(err, "list pod container cpu usage metric by namespaced name failed")
	}
	queryExpression = fmt.Sprintf(`1000 * %s`, queryExpression)
	scope.Debugf("Query to prometheus: queryExpression: %+v, StartTime: %+v, EndTime: %+v, StepTime: %+v", queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
	response, err := prometheusClient.QueryRange(ctx, queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
	if err != nil {
		return nil, errors.Wrap(err, "query prometheus failed")
	} else if response.Status != InternalPromth.StatusSuccess {
		return nil, errors.Errorf("receive error response from prometheus: %s", response.Error)
	}

	entities, err := response.GetEntities()
	if err != nil {
		return nil, errors.Wrap(err, "get prometheus entities")
	}
	cpuUsageEntities := make([]EntityPromthMetric.ContainerCPUUsageMillicoresEntity, len(entities))
	for i, e := range entities {
		cpuUsageEntities[i] = c.newContainerCPUUsageMillicoresEntity(e)
	}

	return cpuUsageEntities, nil
}

func (c ContainerCpuUsageRepository) newContainerCPUUsageMillicoresEntity(e InternalPromth.Entity) EntityPromthMetric.ContainerCPUUsageMillicoresEntity {

	var (
		samples []FormatTypes.Sample
	)

	samples = make([]FormatTypes.Sample, 0)

	for _, value := range e.Values {
		sample := FormatTypes.Sample{
			Timestamp: value.UnixTime,
			Value:     value.SampleValue,
		}
		samples = append(samples, sample)
	}

	return EntityPromthMetric.ContainerCPUUsageMillicoresEntity{
		PrometheusEntity: e,
		Namespace:        e.Labels[ContainerCpuUsagePercentageLabelNamespace],
		PodName:          e.Labels[ContainerCpuUsagePercentageLabelPodName],
		ContainerName:    e.Labels[ContainerCpuUsagePercentageLabelContainerName],
		Samples:          samples,
	}
}
