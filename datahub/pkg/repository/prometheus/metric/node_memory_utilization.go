package metric

import (
	"fmt"
	EntityPromthMetric "github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/metric"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/pkg/errors"
	"time"
)

// NodeMemoryUtilizationRepository Repository to access metric from prometheus
type NodeMemoryUtilizationRepository struct {
	PrometheusConfig InternalPromth.Config
}

// NewNodeMemoryUtilizationRepositoryWithConfig New node cpu utilization percentage repository with prometheus configuration
func NewNodeMemoryUtilizationRepositoryWithConfig(cfg InternalPromth.Config) NodeMemoryUtilizationRepository {
	return NodeMemoryUtilizationRepository{PrometheusConfig: cfg}
}

// ListMetricsByNodeName Provide metrics from response of querying request contain namespace, pod_name and default labels
func (n NodeMemoryUtilizationRepository) ListMetricsByNodeName(nodeName string, options ...DBCommon.Option) ([]InternalPromth.Entity, error) {

	var (
		err error

		prometheusClient *InternalPromth.Prometheus

		nodeMemoryUtilizationMetricName   string
		nodeMemoryUtilizationLabelsString string
		queryExpression                   string

		response InternalPromth.Response

		entities []InternalPromth.Entity
	)

	prometheusClient, err = InternalPromth.NewClient(&n.PrometheusConfig)
	if err != nil {
		return entities, errors.Wrap(err, "list node memory utilization by node name failed")
	}

	opt := DBCommon.NewDefaultOptions()
	for _, option := range options {
		option(&opt)
	}

	nodeMemoryUtilizationMetricName = EntityPromthMetric.NodeMemoryUtilizationMetricName
	nodeMemoryUtilizationLabelsString = n.buildNodeMemoryUtilizationQueryLabelsStringByNodeName(nodeName)

	if nodeMemoryUtilizationLabelsString != "" {
		queryExpression = fmt.Sprintf("%s{%s}", nodeMemoryUtilizationMetricName, nodeMemoryUtilizationLabelsString)
	} else {
		queryExpression = fmt.Sprintf("%s", nodeMemoryUtilizationMetricName)
	}

	stepTimeInSeconds := int64(opt.StepTime.Nanoseconds() / int64(time.Second))
	queryExpression, err = InternalPromth.WrapQueryExpression(queryExpression, opt.AggregateOverTimeFunc, stepTimeInSeconds)
	if err != nil {
		return entities, errors.Wrap(err, "list node memory utilization by node name failed")
	}

	response, err = prometheusClient.QueryRange(queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
	if err != nil {
		return entities, errors.Wrap(err, "list node memory utilization by node name failed")
	} else if response.Status != InternalPromth.StatusSuccess {
		return entities, errors.Errorf("list node memory utilization by node name failed: receive error response from prometheus: %s", response.Error)
	}

	entities, err = response.GetEntities()
	if err != nil {
		return entities, errors.Wrap(err, "list node memory utilization by node name failed")
	}

	return entities, nil
}

func (n NodeMemoryUtilizationRepository) buildNodeMemoryUtilizationQueryLabelsStringByNodeName(nodeName string) string {

	var (
		queryLabelsString = ""
	)

	if nodeName != "" {
		queryLabelsString += fmt.Sprintf(`%s = "%s"`, EntityPromthMetric.NodeMemoryUtilizationLabelNode, nodeName)
	}

	return queryLabelsString
}
