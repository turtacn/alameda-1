package metric

import (
	"fmt"
	EntityPromthNodeMemBytes "github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/nodeMemoryBytesTotal"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/pkg/errors"
	"github.com/containers-ai/alameda/pkg/utils/log"
)

var (
	node_memory_bytes_total_scope = log.RegisterScope("node memory bytes total", "", 0)
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

	node_memory_bytes_total_scope.Infof("node_memory_bytes_total_scope metric-ListMetricsByNodeName input nodename %s", nodeName)
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
		node_memory_bytes_total_scope.Errorf("node_memory_bytes_total_scope metric-ListMetricsByNodeName error %v", err)
		return entities, errors.Wrap(err, "list node memory utilization by node name failed")
	}

	opt := DBCommon.NewDefaultOptions()
	for _, option := range options {
		option(&opt)
	}

	nodeMemoryBytesTotalMetricName = EntityPromthNodeMemBytes.MetricName
	nodeMemoryBytesTotalQueryLabelsString = n.buildNodeMemoryBytesTotalQueryLabelsStringByNodeName(nodeName)

	if nodeMemoryBytesTotalQueryLabelsString != "" {
		queryExpression = fmt.Sprintf("%s{%s}", nodeMemoryBytesTotalMetricName, nodeMemoryBytesTotalQueryLabelsString)
	} else {
		queryExpression = fmt.Sprintf("%s", nodeMemoryBytesTotalMetricName)
	}

	response, err = prometheusClient.QueryRange(queryExpression, opt.StartTime, opt.EndTime, opt.StepTime)
	if err != nil {
		node_memory_bytes_total_scope.Errorf("node_memory_bytes_total_scope metric-ListMetricsByNodeName error %v", err)
		return entities, errors.Wrap(err, "list node memory bytes total by node name failed")
	} else if response.Status != InternalPromth.StatusSuccess {
		node_memory_bytes_total_scope.Errorf("node_memory_bytes_total_scope metric-ListMetricsByNodeName error resonse status not success")
		return entities, errors.Errorf("list node memory bytes total by node name failed: receive error response from prometheus: %s", response.Error)
	}

	entities, err = response.GetEntities()
	if err != nil {
		node_memory_bytes_total_scope.Errorf("node_memory_bytes_total_scope metric-ListMetricsByNodeName error %v", err)
		return entities, errors.Wrap(err, "list node memory bytes total by node name failed")
	}
	node_memory_bytes_total_scope.Errorf("node_memory_bytes_total_scope metric-ListMetricsByNodeName return %d %v", len(entities), &entities)
	return entities, nil
}

func (n NodeMemoryBytesTotalRepository) buildNodeMemoryBytesTotalQueryLabelsStringByNodeName(nodeName string) string {

	var (
		queryLabelsString = ""
	)

	if nodeName != "" {
		queryLabelsString += fmt.Sprintf(`%s = "%s"`, EntityPromthNodeMemBytes.NodeLabel, nodeName)
	}

	return queryLabelsString
}
