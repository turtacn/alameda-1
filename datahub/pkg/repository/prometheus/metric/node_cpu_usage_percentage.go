package metric

import (
	"fmt"
	EntityPromthNodeCpu "github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/nodeCPUUsagePercentage"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/pkg/errors"
	"time"
	"github.com/containers-ai/alameda/pkg/utils/log"
)


var (
	node_cpu_usage_percentage_scope = log.RegisterScope("node cpu usage  percentage", "", 0)
)


// NodeCPUUsagePercentageRepository Repository to access metric node:node_cpu_utilisation:avg1m from prometheus
type NodeCPUUsagePercentageRepository struct {
	PrometheusConfig InternalPromth.Config
}

// NewNodeCPUUsagePercentageRepositoryWithConfig New node cpu usage percentage repository with prometheus configuration
func NewNodeCPUUsagePercentageRepositoryWithConfig(cfg InternalPromth.Config) NodeCPUUsagePercentageRepository {
	return NodeCPUUsagePercentageRepository{PrometheusConfig: cfg}
}

// ListMetricsByPodNamespacedName Provide metrics from response of querying request contain namespace, pod_name and default labels
func (n NodeCPUUsagePercentageRepository) ListMetricsByNodeName(nodeName string, options ...DBCommon.Option) ([]InternalPromth.Entity, error) {
	node_cpu_usage_percentage_scope.Infof("node_cpu_usage_percentage_scope metric-ListMetricsByNodeName input {%s, %v}", nodeName, options)
	var (
		err error

		prometheusClient *InternalPromth.Prometheus

		queryLabelsString string

		queryExpressionSum string
		queryExpressionAvg string

		response InternalPromth.Response

		entities []InternalPromth.Entity
	)

	prometheusClient, err = InternalPromth.NewClient(&n.PrometheusConfig)
	if err != nil {
		node_cpu_usage_percentage_scope.Errorf("node_cpu_usage_percentage_scope metric-ListMetricsByNodeName error %v", err)
		return entities, errors.Wrap(err, "list node cpu usage metrics by node name failed")
	}

	opt := DBCommon.NewDefaultOptions()
	for _, option := range options {
		option(&opt)
	}

	//metricName = EntityPromthNodeCpu.MetricName
	metricNameSum := EntityPromthNodeCpu.MetricNameSum
	metricNameAvg := EntityPromthNodeCpu.MetricNameAvg

	queryLabelsString = n.buildQueryLabelsStringByNodeName(nodeName)

	if queryLabelsString != "" {
		queryExpressionSum = fmt.Sprintf("%s{%s}", metricNameSum, queryLabelsString)
		queryExpressionAvg = fmt.Sprintf("%s{%s}", metricNameAvg, queryLabelsString)
	} else {
		queryExpressionSum = fmt.Sprintf("%s", metricNameSum)
		queryExpressionAvg = fmt.Sprintf("%s", metricNameAvg)
	}

	stepTimeInSeconds := int64(opt.StepTime.Nanoseconds() / int64(time.Second))
	queryExpressionSum, err = InternalPromth.WrapQueryExpression(queryExpressionSum, opt.AggregateOverTimeFunc, stepTimeInSeconds)
	if err != nil {
		node_cpu_usage_percentage_scope.Errorf("node_cpu_usage_percentage_scope metric-ListMetricsByNodeName error %v", err)
		return entities, errors.Wrap(err, "list node cpu usage metrics by node name failed")
	}
	queryExpressionAvg, err = InternalPromth.WrapQueryExpression(queryExpressionAvg, opt.AggregateOverTimeFunc, stepTimeInSeconds)
	if err != nil {
		node_cpu_usage_percentage_scope.Errorf("node_cpu_usage_percentage_scope metric-ListMetricsByNodeName error %v", err)
		return entities, errors.Wrap(err, "list node cpu usage metrics by node name failed")
	}

	queryExpression := fmt.Sprintf("1000 * %s * %s", queryExpressionSum, queryExpressionAvg)

	response, err = prometheusClient.QueryRange(queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
	if err != nil {
		node_cpu_usage_percentage_scope.Errorf("node_cpu_usage_percentage_scope metric-ListMetricsByNodeName error %v", err)
		return entities, errors.Wrap(err, "list node cpu usage metrics by node name failed")
	} else if response.Status != InternalPromth.StatusSuccess {
		node_cpu_usage_percentage_scope.Errorf("node_cpu_usage_percentage_scope metric-ListMetricsByNodeName response status != prometheus success status %v")
		return entities, errors.Errorf("list node cpu usage metrics by node name failed: receive error response from prometheus: %s", response.Error)
	}

	entities, err = response.GetEntities()
	if err != nil {
		node_cpu_usage_percentage_scope.Errorf("node_cpu_usage_percentage_scope metric-ListMetricsByNodeName error %v", err)
		return entities, errors.Wrap(err, "list node cpu usage metrics by node name failed")
	}
	node_cpu_usage_percentage_scope.Infof("node_cpu_usage_percentage_scope metric-ListMetricsByNodeName return %d  %v", len(entities) , &entities)
	return entities, nil
}

func (n NodeCPUUsagePercentageRepository) buildQueryLabelsStringByNodeName(nodeName string) string {

	var (
		queryLabelsString = ""
	)

	if nodeName != "" {
		//queryLabelsString += fmt.Sprintf(`%s = "%s"`, EntityPromthNodeCpu.NodeLabel, nodeName)
	}

	return queryLabelsString
}
