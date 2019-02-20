package metric

import (
	"fmt"
	"time"

	"github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/nodeMemoryAvailableBytes"
	"github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/nodeMemoryBytesTotal"
	"github.com/containers-ai/alameda/datahub/pkg/repository/prometheus"
	"github.com/pkg/errors"
)

// NodeMemoryUsageBytesRepository Repository to access metric from prometheus
type NodeMemoryUsageBytesRepository struct {
	PrometheusConfig prometheus.Config
}

// NewNodeMemoryUsageBytesRepositoryWithConfig New node cpu usage percentage repository with prometheus configuration
func NewNodeMemoryUsageBytesRepositoryWithConfig(cfg prometheus.Config) NodeMemoryUsageBytesRepository {
	return NodeMemoryUsageBytesRepository{PrometheusConfig: cfg}
}

// ListMetricsByPodNamespacedName Provide metrics from response of querying request contain namespace, pod_name and default labels
func (n NodeMemoryUsageBytesRepository) ListMetricsByNodeName(nodeName string, startTime, endTime *time.Time, stepTime *time.Duration) ([]prometheus.Entity, error) {

	var (
		err error

		prometheusClient *prometheus.Prometheus

		nodeMemoryBytesTotalMetricName            string
		nodeMemoryAvailableBytesMetricName        string
		nodeMemoryBytesTotalQueryLabelsString     string
		nodeMemoryAvailableBytesQueryLabelsString string
		queryExpression                           string

		response prometheus.Response

		entities []prometheus.Entity
	)

	prometheusClient, err = prometheus.New(n.PrometheusConfig)
	if err != nil {
		return entities, errors.Wrap(err, "list node memory usage by node name failed")
	}

	nodeMemoryBytesTotalMetricName = nodeMemoryBytesTotal.MetricName
	nodeMemoryAvailableBytesMetricName = nodeMemoryAvailableBytes.MetricName
	nodeMemoryBytesTotalQueryLabelsString = n.buildNodeMemoryBytesTotalQueryLabelsStringByNodeName(nodeName)
	nodeMemoryAvailableBytesQueryLabelsString = n.buildNodeMemoryAvailableBytesQueryLabelsStringByNodeName(nodeName)

	if nodeMemoryBytesTotalQueryLabelsString != "" && nodeMemoryAvailableBytesQueryLabelsString != "" {
		queryExpression = fmt.Sprintf("%s{%s} - %s{%s}", nodeMemoryBytesTotalMetricName, nodeMemoryBytesTotalQueryLabelsString, nodeMemoryAvailableBytesMetricName, nodeMemoryAvailableBytesQueryLabelsString)
	} else {
		queryExpression = fmt.Sprintf("%s - %s", nodeMemoryBytesTotalMetricName, nodeMemoryAvailableBytesMetricName)
	}

	response, err = prometheusClient.QueryRange(queryExpression, startTime, endTime, stepTime)
	if err != nil {
		return entities, errors.Wrap(err, "list node memory usage by node name failed")
	} else if response.Status != prometheus.StatusSuccess {
		return entities, errors.Errorf("list node memory usage by node name failed: receive error response from prometheus: %s", response.Error)
	}

	entities, err = response.GetEntitis()
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
		queryLabelsString += fmt.Sprintf(`%s = "%s"`, nodeMemoryBytesTotal.NodeLabel, nodeName)
	}

	return queryLabelsString
}

func (n NodeMemoryUsageBytesRepository) buildNodeMemoryAvailableBytesQueryLabelsStringByNodeName(nodeName string) string {

	var (
		queryLabelsString = ""
	)

	if nodeName != "" {
		queryLabelsString += fmt.Sprintf(`%s = "%s"`, nodeMemoryAvailableBytes.NodeLabel, nodeName)
	}

	return queryLabelsString
}
