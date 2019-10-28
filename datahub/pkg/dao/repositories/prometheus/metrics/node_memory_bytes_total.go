package metrics

import (
	"fmt"
	EntityPromthMetric "github.com/containers-ai/alameda/datahub/pkg/dao/entities/prometheus/metrics"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/pkg/errors"
)

// NodeMemoryBytesTotalRepository Repository to access metric from prometheus
type NodeMemoryBytesTotalRepository struct {
	PrometheusConfig InternalPromth.Config
}

// NewNodeMemoryBytesTotalRepositoryWithConfig New node cpu utilization percentage repository with prometheus configuration
func NewNodeMemoryBytesTotalRepositoryWithConfig(cfg InternalPromth.Config) NodeMemoryBytesTotalRepository {
	return NodeMemoryBytesTotalRepository{PrometheusConfig: cfg}
}

func (n NodeMemoryBytesTotalRepository) ListMetricsByNodeName(nodeName string, options ...DBCommon.Option) ([]InternalPromth.Entity, error) {

	var (
		err error

		prometheusClient *InternalPromth.Prometheus

		nodeMemoryBytesTotalMetricName        string
		nodeMemoryBytesTotalQueryLabelsString string
		queryExpression                       string

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

	nodeMemoryBytesTotalMetricName = EntityPromthMetric.NodeMemoryBytesTotalMetricName
	nodeMemoryBytesTotalQueryLabelsString = n.buildNodeMemoryBytesTotalQueryLabelsStringByNodeName(nodeName)

	if nodeMemoryBytesTotalQueryLabelsString != "" {
		queryExpression = fmt.Sprintf("%s{%s}", nodeMemoryBytesTotalMetricName, nodeMemoryBytesTotalQueryLabelsString)
	} else {
		queryExpression = fmt.Sprintf("%s", nodeMemoryBytesTotalMetricName)
	}

	response, err = prometheusClient.QueryRange(queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
	if err != nil {
		return entities, errors.Wrap(err, "list node memory bytes total by node name failed")
	} else if response.Status != InternalPromth.StatusSuccess {
		return entities, errors.Errorf("list node memory bytes total by node name failed: receive error response from prometheus: %s", response.Error)
	}

	entities, err = response.GetEntities()
	if err != nil {
		return entities, errors.Wrap(err, "list node memory bytes total by node name failed")
	}

	return entities, nil
}

func (n NodeMemoryBytesTotalRepository) buildNodeMemoryBytesTotalQueryLabelsStringByNodeName(nodeName string) string {

	var (
		queryLabelsString = ""
	)

	if nodeName != "" {
		queryLabelsString += fmt.Sprintf(`%s = "%s"`, EntityPromthMetric.NodeMemoryBytesTotalLabelNode, nodeName)
	}

	return queryLabelsString
}
