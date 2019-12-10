package prometheus

import (
	"context"

	DaoClusterStatusTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	RepoPromthMetric "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/prometheus/metrics"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type NodeMetrics struct {
	PrometheusConfig InternalPromth.Config

	nodeDAO DaoClusterStatusTypes.NodeDAO

	clusterUID string
}

// NewNodeMetricsWithConfig Constructor of prometheus node metric dao
func NewNodeMetricsWithConfig(config InternalPromth.Config, nodeDAO DaoClusterStatusTypes.NodeDAO, clusterUID string) DaoMetricTypes.NodeMetricsDAO {
	return &NodeMetrics{
		PrometheusConfig: config,

		nodeDAO: nodeDAO,

		clusterUID: clusterUID,
	}
}

// CreateMetrics Method implementation of NodeMetricsDAO
func (p *NodeMetrics) CreateMetrics(ctx context.Context, metrics DaoMetricTypes.NodeMetricMap) error {
	return errors.New("create metrics to prometheus is not supported")
}

// ListMetrics Method implementation of NodeMetricsDAO
func (p *NodeMetrics) ListMetrics(ctx context.Context, req DaoMetricTypes.ListNodeMetricsRequest) (DaoMetricTypes.NodeMetricMap, error) {

	options := []DBCommon.Option{
		DBCommon.StartTime(req.StartTime),
		DBCommon.EndTime(req.EndTime),
		DBCommon.StepTime(req.StepTime),
		DBCommon.AggregateOverTimeFunc(req.AggregateOverTimeFunction),
	}

	nodeMetas, err := p.listNodeMetasFromRequest(ctx, req)
	if err != nil {
		return DaoMetricTypes.NodeMetricMap{}, errors.Wrap(err, "list node metadatas from request failed")
	}
	nodeMetas = filterObjectMetaByClusterUID(p.clusterUID, nodeMetas)
	if len(nodeMetas) == 0 {
		return DaoMetricTypes.NodeMetricMap{}, nil
	}
	metricMap, err := p.getNodeMetricMapByObjectMetas(ctx, nodeMetas, options...)
	if err != nil {
		return DaoMetricTypes.NodeMetricMap{}, errors.Wrap(err, "get node metricMap failed")
	}
	metricMap.SortByTimestamp(req.QueryCondition.TimestampOrder)
	metricMap.Limit(req.QueryCondition.Limit)
	return metricMap, nil
}

func (p *NodeMetrics) listNodeMetasFromRequest(ctx context.Context, req DaoMetricTypes.ListNodeMetricsRequest) ([]metadata.ObjectMeta, error) {

	// Generate list resource nodes request
	listNodesReq := DaoClusterStatusTypes.NewListNodesRequest()
	for index := range req.ObjectMetas {
		listNodesReq.ObjectMeta = append(listNodesReq.ObjectMeta, &req.ObjectMetas[index])
	}

	nodes, err := p.nodeDAO.ListNodes(listNodesReq)
	if err != nil {
		return nil, errors.Wrap(err, "list nodes by cluster metadata failed")
	}
	objectMetas := make([]metadata.ObjectMeta, len(nodes))
	for i, node := range nodes {
		objectMetas[i] = *node.ObjectMeta
	}
	return objectMetas, nil
}

func (p *NodeMetrics) getNodeMetricMapByObjectMetas(ctx context.Context, nodeMetas []metadata.ObjectMeta, options ...DBCommon.Option) (DaoMetricTypes.NodeMetricMap, error) {

	nodeMap := make(map[string]metadata.ObjectMeta)
	names := make([]string, len(nodeMetas))
	for i, nodeMeta := range nodeMetas {
		names[i] = nodeMeta.Name
		nodeMap[nodeMeta.Name] = nodeMeta
	}

	metricChan := make(chan DaoMetricTypes.NodeMetric)
	producerWG := errgroup.Group{}
	// Query cpu metrics
	producerWG.Go(func() error {

		nodeCPUUsageRepo := RepoPromthMetric.NewNodeCPUUsageRepositoryWithConfig(p.PrometheusConfig)
		nodeCPUUsageEntities, err := nodeCPUUsageRepo.ListNodeCPUUsageMillicoresEntitiesByNodeNames(ctx, names, options...)
		if err != nil {
			return errors.Wrap(err, "list node cpu usage metrics failed")
		}

		for _, e := range nodeCPUUsageEntities {
			nodeMetric := e.NodeMetric()
			nodeMetric.ObjectMeta = nodeMap[e.NodeName]
			metricChan <- nodeMetric
		}

		return nil
	})
	// Query memory metrics
	producerWG.Go(func() error {

		nodeMemoryUsageRepo := RepoPromthMetric.NewNodeMemoryUsageRepositoryWithConfig(p.PrometheusConfig)
		nodeMemoryUsageEntities, err := nodeMemoryUsageRepo.ListNodeMemoryBytesUsageEntitiesByNodeNames(ctx, names, options...)
		if err != nil {
			return errors.Wrap(err, "list node memory usage metrics failed")
		}

		for _, e := range nodeMemoryUsageEntities {
			nodeMetric := e.NodeMetric()
			nodeMetric.ObjectMeta = nodeMap[e.NodeName]
			metricChan <- nodeMetric
		}

		return nil
	})

	metricMap := DaoMetricTypes.NewNodeMetricMap()
	consumerWG := errgroup.Group{}
	consumerWG.Go(func() error {

		for m := range metricChan {
			copyM := m
			metricMap.AddNodeMetric(&copyM)
		}
		return nil
	})

	err := producerWG.Wait()
	close(metricChan)
	if err != nil {
		return DaoMetricTypes.NodeMetricMap{}, err
	}

	consumerWG.Wait()
	for _, nodeMeta := range nodeMetas {
		if metric, exist := metricMap.MetricMap[nodeMeta]; !exist || metric == nil {
			metricMap.MetricMap[nodeMeta] = &DaoMetricTypes.NodeMetric{
				ObjectMeta: nodeMeta,
			}
		}
	}
	return metricMap, nil
}
