package metric

import (
	"fmt"
	EntityPromthNodeMemBytes "github.com/containers-ai/alameda/datapipe/pkg/entities/prometheus/nodeMemoryBytesTotal"
	EntityPromthNodeMemUtilization "github.com/containers-ai/alameda/datapipe/pkg/entities/prometheus/nodeMemoryUtilization"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/containers-ai/alameda/pkg/utils/log"
	"github.com/pkg/errors"
	"time"
)

var (
	scope = log.RegisterScope("node memory usage bytes", "node memory usage bytes", 0)
)

// NodeMemoryUsageBytesRepository Repository to access metric from prometheus
type NodeMemoryUsageBytesRepository struct {
	PrometheusConfig InternalPromth.Config
}

// NewNodeMemoryUsageBytesRepositoryWithConfig New node cpu usage percentage repository with prometheus configuration
func NewNodeMemoryUsageBytesRepositoryWithConfig(cfg InternalPromth.Config) NodeMemoryUsageBytesRepository {
	return NodeMemoryUsageBytesRepository{PrometheusConfig: cfg}
}

// ListMetricsByNodeName Provide metrics from response of querying request contain namespace, pod_name and default labels
func (n NodeMemoryUsageBytesRepository) ListMetricsByNodeName(nodeName string, options ...DBCommon.Option) ([]InternalPromth.Entity, error) {

	var (
		err error

		prometheusClient *InternalPromth.Prometheus

		nodeMemoryBytesTotalQueryExpression    string
		nodeMemoryBytesTotalMetricName         string
		nodeMemoryBytesTotalQueryLabelsString  string
		nodeMemoryUtilizationQueryExpression   string
		nodeMemoryUtilizationMetricName        string
		nodeMemoryUtilizationQueryLabelsString string
		queryExpression                        string

		response InternalPromth.Response

		entities []InternalPromth.Entity
	)

	prometheusClient, err = InternalPromth.NewClient(&n.PrometheusConfig)
	if err != nil {
		return entities, errors.Wrap(err, "list node memory usage by node name failed")
	}

	opt := DBCommon.NewDefaultOptions()
	for _, option := range options {
		option(&opt)
	}
	stepTimeInSeconds := int64(opt.StepTime.Nanoseconds() / int64(time.Second))

	nodeMemoryBytesTotalMetricName = EntityPromthNodeMemBytes.MetricName
	nodeMemoryBytesTotalQueryLabelsString = n.buildNodeMemoryBytesTotalQueryLabelsStringByNodeName(nodeName)
	if nodeMemoryBytesTotalQueryLabelsString != "" {
		nodeMemoryBytesTotalQueryExpression = fmt.Sprintf("%s{%s}", nodeMemoryBytesTotalMetricName, nodeMemoryBytesTotalQueryLabelsString)
	} else {
		nodeMemoryBytesTotalQueryExpression = fmt.Sprintf("%s", nodeMemoryBytesTotalMetricName)
	}
	nodeMemoryBytesTotalQueryExpression, err = InternalPromth.WrapQueryExpression(nodeMemoryBytesTotalQueryExpression, opt.AggregateOverTimeFunc, stepTimeInSeconds)
	if err != nil {
		return entities, errors.Wrap(err, "list node memory usage metrics by node name failed")
	}

	nodeMemoryUtilizationMetricName = EntityPromthNodeMemUtilization.MetricName
	nodeMemoryUtilizationQueryLabelsString = n.buildNodeMemoryUtilizationQueryLabelsStringByNodeName(nodeName)
	if nodeMemoryUtilizationQueryLabelsString != "" {
		nodeMemoryUtilizationQueryExpression = fmt.Sprintf("%s{%s}", nodeMemoryUtilizationMetricName, nodeMemoryUtilizationQueryLabelsString)
	} else {
		nodeMemoryUtilizationQueryExpression = fmt.Sprintf("%s", nodeMemoryUtilizationMetricName)
	}
	nodeMemoryUtilizationQueryExpression, err = InternalPromth.WrapQueryExpression(nodeMemoryUtilizationQueryExpression, opt.AggregateOverTimeFunc, stepTimeInSeconds)
	if err != nil {
		return entities, errors.Wrap(err, "list node memory usage metrics by node name failed")
	}

	queryExpression = fmt.Sprintf("%s * %s", nodeMemoryBytesTotalQueryExpression, nodeMemoryUtilizationQueryExpression)

	response, err = prometheusClient.QueryRange(queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
	if err != nil {
		return entities, errors.Wrap(err, "list node memory bytes total by node name failed")
	} else if response.Status != InternalPromth.StatusSuccess {
		return entities, errors.Errorf("list node memory bytes total by node name failed: receive error response from prometheus: %s", response.Error)
	}

	entities, err = response.GetEntities()
	if err != nil {
		return entities, errors.Wrap(err, "list node memory usage by node name failed")
	}

	return entities, nil
}

func (n NodeMemoryUsageBytesRepository) buildNodeMemoryBytesTotalQueryLabelsStringByNodeName(nodeName string) string {

	var (
		queryLabelsString = ""
	)

	if nodeName != "" {
		queryLabelsString += fmt.Sprintf(`%s = "%s"`, EntityPromthNodeMemBytes.NodeLabel, nodeName)
	}

	return queryLabelsString
}

func (n NodeMemoryUsageBytesRepository) buildNodeMemoryUtilizationQueryLabelsStringByNodeName(nodeName string) string {

	var (
		queryLabelsString = ""
	)

	if nodeName != "" {
		queryLabelsString += fmt.Sprintf(`%s = "%s"`, EntityPromthNodeMemUtilization.NodeLabel, nodeName)
	}

	return queryLabelsString
}
