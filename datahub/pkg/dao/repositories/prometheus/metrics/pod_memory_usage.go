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

// PodMemoryUsageRepository Repository to access metric container_memory_usage_bytes from prometheus
type PodMemoryUsageRepository struct {
	PrometheusConfig InternalPromth.Config
}

// NewPodMemoryUsageRepositoryWithConfig New pod memory usage bytes repository with prometheus configuration
func NewPodMemoryUsageRepositoryWithConfig(cfg InternalPromth.Config) PodMemoryUsageRepository {
	return PodMemoryUsageRepository{PrometheusConfig: cfg}
}

func (c PodMemoryUsageRepository) ListPodMemoryUsageBytesEntityBySummingPodMetrics(ctx context.Context, namespace string, podNames []string, options ...DBCommon.Option) ([]EntityPromthMetric.PodMemoryUsageBytesEntity, error) {
	// Example of expression to query prometheus
	// sum(container_memory_usage_bytes{pod_name!="",container_name!="",container_name!="POD",namespace="@n1",pod_name=~"@p1|@p2"})

	prometheusClient, err := InternalPromth.NewClient(&c.PrometheusConfig)
	if err != nil {
		return nil, errors.Wrap(err, "new prometheus client failed")
	}

	opt := DBCommon.NewDefaultOptions()
	for _, option := range options {
		option(&opt)
	}

	queryLabelsString := c.buildDefaultQueryLabelsString()
	queryLabelsString += fmt.Sprintf(`, %s = "%s"`, ContainerMemoryUsageBytesLabelNamespace, namespace)
	names := ""
	for _, name := range podNames {
		names += fmt.Sprintf("%s|", name)
	}
	if names != "" {
		names = strings.TrimSuffix(names, "|")
		queryLabelsString += fmt.Sprintf(`,%s =~ "%s"`, ContainerCpuUsagePercentageLabelPodName, names)
	}
	queryExpression := fmt.Sprintf(`%s{%s}`, ContainerMemoryUsageBytesMetricName, queryLabelsString)
	stepTimeInSeconds := int64(opt.StepTime.Nanoseconds() / int64(time.Second))
	queryExpression, err = InternalPromth.WrapQueryExpression(queryExpression, opt.AggregateOverTimeFunc, stepTimeInSeconds)
	if err != nil {
		return nil, errors.Wrap(err, "wrap query expression failed")
	}
	queryExpression = fmt.Sprintf(`sum(%s)`, queryExpression)

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
	podMemoryUsageBytesEntities := make([]EntityPromthMetric.PodMemoryUsageBytesEntity, len(entities))
	for i, entity := range entities {
		e := c.newPodMemoryUsageBytesEntity(entity)
		podMemoryUsageBytesEntities[i] = e
	}

	return podMemoryUsageBytesEntities, nil
}

func (c PodMemoryUsageRepository) buildDefaultQueryLabelsString() string {
	// sum(container_memory_usage_bytes{pod_name!="",container_name!="",container_name!="POD",namespace="@n1",pod_name=~"@p1|@p2"})

	var queryLabelsString = ""
	queryLabelsString += fmt.Sprintf(`%s != "",`, ContainerMemoryUsageBytesLabelPodName)
	queryLabelsString += fmt.Sprintf(`%s != "",`, ContainerMemoryUsageBytesLabelContainerName)
	queryLabelsString += fmt.Sprintf(`%s != "POD"`, ContainerMemoryUsageBytesLabelContainerName)
	return queryLabelsString
}

func (c PodMemoryUsageRepository) newPodMemoryUsageBytesEntity(e InternalPromth.Entity) EntityPromthMetric.PodMemoryUsageBytesEntity {

	samples := make([]FormatTypes.Sample, len(e.Values))
	for i, value := range e.Values {
		samples[i] = FormatTypes.Sample{
			Timestamp: value.UnixTime,
			Value:     value.SampleValue,
		}
	}
	return EntityPromthMetric.PodMemoryUsageBytesEntity{
		NamespaceName: e.Labels[ContainerCpuUsagePercentageLabelNamespace],
		PodName:       e.Labels[ContainerCpuUsagePercentageLabelPodName],
		Samples:       samples,
	}
}
