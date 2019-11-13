package metrics

import (
	"context"
	"fmt"
	"strings"

	EntityPromthMetric "github.com/containers-ai/alameda/datahub/pkg/dao/entities/prometheus/metrics"
	FormatTypes "github.com/containers-ai/alameda/datahub/pkg/formatconversion/types"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/pkg/errors"
	"time"
)

type NodeMemoryUsageRepository struct {
	PrometheusConfig InternalPromth.Config
}

func NewNodeMemoryUsageRepositoryWithConfig(cfg InternalPromth.Config) NodeMemoryUsageRepository {
	return NodeMemoryUsageRepository{PrometheusConfig: cfg}
}

func (n NodeMemoryUsageRepository) ListNodeMemoryBytesUsageEntitiesByNodeNames(ctx context.Context, nodeNames []string, options ...DBCommon.Option) ([]EntityPromthMetric.NodeMemoryBytesUsageEntity, error) {

	prometheusClient, err := InternalPromth.NewClient(&n.PrometheusConfig)
	if err != nil {
		return nil, errors.Wrap(err, "new prometheus client failed")
	}

	opt := DBCommon.NewDefaultOptions()
	for _, option := range options {
		option(&opt)
	}
	stepTimeInSeconds := int64(opt.StepTime.Nanoseconds() / int64(time.Second))

	names := ""
	for _, nodeName := range nodeNames {
		names += fmt.Sprintf("%s|", nodeName)
	}
	names = strings.TrimSuffix(names, "|")

	nodeMemoryBytesTotalQueryLabelsString := ""
	if names != "" {
		nodeMemoryBytesTotalQueryLabelsString = fmt.Sprintf(`%s =~ "%s"`, NodeMemoryBytesTotalLabelNode, names)
	}
	nodeMemoryBytesTotalQueryExpression := fmt.Sprintf("%s{%s}", NodeMemoryBytesTotalMetricName, nodeMemoryBytesTotalQueryLabelsString)
	nodeMemoryBytesTotalQueryExpression, err = InternalPromth.WrapQueryExpression(nodeMemoryBytesTotalQueryExpression, opt.AggregateOverTimeFunc, stepTimeInSeconds)
	if err != nil {
		return nil, errors.Wrap(err, "wrap query expression failed")
	}

	nodeMemoryUtilizationQueryLabelsString := ""
	if names != "" {
		nodeMemoryUtilizationQueryLabelsString = fmt.Sprintf(`%s =~ "%s"`, NodeMemoryBytesTotalLabelNode, names)
	}
	nodeMemoryUtilizationQueryExpression := fmt.Sprintf("%s{%s}", NodeMemoryUtilizationMetricName, nodeMemoryUtilizationQueryLabelsString)
	nodeMemoryUtilizationQueryExpression, err = InternalPromth.WrapQueryExpression(nodeMemoryUtilizationQueryExpression, opt.AggregateOverTimeFunc, stepTimeInSeconds)
	if err != nil {
		return nil, errors.Wrap(err, "wrap query expression failed")
	}

	queryExpression := fmt.Sprintf("%s * %s", nodeMemoryBytesTotalQueryExpression, nodeMemoryUtilizationQueryExpression)
	scope.Debugf("Query to prometheus: queryExpression: %+v, StartTime: %+v, EndTime: %+v, StepTime: %+v", queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
	response, err := prometheusClient.QueryRange(ctx, queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
	if err != nil {
		return nil, errors.Wrap(err, "query prometheus failed")
	} else if response.Status != InternalPromth.StatusSuccess {
		return nil, errors.Errorf("receive error response from prometheus: %s", response.Error)
	}

	entities, err := response.GetEntities()
	if err != nil {
		return nil, errors.Wrap(err, "get prometheus entitiesfailed")
	}
	memoryUsageEntities := make([]EntityPromthMetric.NodeMemoryBytesUsageEntity, len(entities))
	for i, e := range entities {
		memoryUsageEntities[i] = n.newNodeMemoryBytesUsageEntity(e)
	}

	return memoryUsageEntities, nil
}

func (n NodeMemoryUsageRepository) ListSumOfNodeMetricsByNodeNames(ctx context.Context, nodeNames []string, options ...DBCommon.Option) ([]EntityPromthMetric.NodeMemoryBytesUsageEntity, error) {
	// Example of expression to query prometheus
	// sum (node:node_memory_bytes_total:sum * node:node_memory_utilisation_2:)

	prometheusClient, err := InternalPromth.NewClient(&n.PrometheusConfig)
	if err != nil {
		return nil, errors.Wrap(err, "new prometheus client failed")
	}

	opt := DBCommon.NewDefaultOptions()
	for _, option := range options {
		option(&opt)
	}
	stepTimeInSeconds := int64(opt.StepTime.Nanoseconds() / int64(time.Second))

	names := ""
	for _, name := range nodeNames {
		names += fmt.Sprintf("%s|", name)
	}
	queryLabelsString := ""
	if names != "" {
		names = strings.TrimSuffix(names, "|")
		queryLabelsString = fmt.Sprintf(`%s=~"%s"`, NodeCpuUsagePercentageLabelNode, names)
	}

	nodeMemoryBytesTotalQueryExpression := fmt.Sprintf("%s{%s}", NodeMemoryBytesTotalMetricName, queryLabelsString)
	nodeMemoryBytesTotalQueryExpression, err = InternalPromth.WrapQueryExpression(nodeMemoryBytesTotalQueryExpression, opt.AggregateOverTimeFunc, stepTimeInSeconds)
	if err != nil {
		return nil, errors.Wrap(err, "wrap query expression failed")
	}

	nodeMemoryUtilizationQueryExpression := fmt.Sprintf("%s{%s}", NodeMemoryUtilizationMetricName, queryLabelsString)
	nodeMemoryUtilizationQueryExpression, err = InternalPromth.WrapQueryExpression(nodeMemoryUtilizationQueryExpression, opt.AggregateOverTimeFunc, stepTimeInSeconds)
	if err != nil {
		return nil, errors.Wrap(err, "wrap query expression failed")
	}

	queryExpression := fmt.Sprintf("sum(%s * %s)", nodeMemoryBytesTotalQueryExpression, nodeMemoryUtilizationQueryExpression)
	scope.Debugf("Query to prometheus: queryExpression: %+v, StartTime: %+v, EndTime: %+v, StepTime: %+v", queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
	response, err := prometheusClient.QueryRange(ctx, queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
	if err != nil {
		return nil, errors.Wrap(err, "query prometheus failed")
	} else if response.Status != InternalPromth.StatusSuccess {
		return nil, errors.Errorf("query prometheus failed: receive error response from prometheus: %s", response.Error)
	}

	entities, err := response.GetEntities()
	if err != nil {
		return nil, errors.Wrap(err, "get prometheus entities failed")
	}
	nodeMemoryBytesUsageEntities := make([]EntityPromthMetric.NodeMemoryBytesUsageEntity, len(entities))
	for i, entity := range entities {
		e := n.newNodeMemoryBytesUsageEntity(entity)
		nodeMemoryBytesUsageEntities[i] = e
	}

	return nodeMemoryBytesUsageEntities, nil
}

func (n NodeMemoryUsageRepository) newNodeMemoryBytesUsageEntity(e InternalPromth.Entity) EntityPromthMetric.NodeMemoryBytesUsageEntity {

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

	return EntityPromthMetric.NodeMemoryBytesUsageEntity{
		NodeName: e.Labels[NodeMemoryBytesUsageLabelNode],
		Samples:  samples,
	}
}
