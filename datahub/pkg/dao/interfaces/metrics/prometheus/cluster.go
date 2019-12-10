package prometheus

import (
	"context"

	EntityPromthMetric "github.com/containers-ai/alameda/datahub/pkg/dao/entities/prometheus/metrics"
	DaoClusterStatusTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	RepoPromthMetric "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/prometheus/metrics"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type ClusterMetrics struct {
	PrometheusConfig InternalPromth.Config

	clusterStatusDAO DaoClusterStatusTypes.ClusterDAO
	nodeDAO          DaoClusterStatusTypes.NodeDAO

	clusterUID string
}

// NewClusterMetricsWithConfig Constructor of prometheus namespace metric dao
func NewClusterMetricsWithConfig(config InternalPromth.Config, clusterStatusDAO DaoClusterStatusTypes.ClusterDAO, nodeDAO DaoClusterStatusTypes.NodeDAO, clusterUID string) DaoMetricTypes.ClusterMetricsDAO {
	return &ClusterMetrics{
		PrometheusConfig: config,

		clusterStatusDAO: clusterStatusDAO,
		nodeDAO:          nodeDAO,

		clusterUID: clusterUID,
	}
}

func (p ClusterMetrics) CreateMetrics(ctx context.Context, m DaoMetricTypes.ClusterMetricMap) error {
	return errors.New("not implemented")
}

func (p ClusterMetrics) ListMetrics(ctx context.Context, req DaoMetricTypes.ListClusterMetricsRequest) (DaoMetricTypes.ClusterMetricMap, error) {

	options := []DBCommon.Option{
		DBCommon.StartTime(req.StartTime),
		DBCommon.EndTime(req.EndTime),
		DBCommon.StepTime(req.StepTime),
		DBCommon.AggregateOverTimeFunc(req.AggregateOverTimeFunction),
	}
	metas, err := p.listClusterMetasFromRequest(ctx, req)
	if err != nil {
		return DaoMetricTypes.ClusterMetricMap{}, errors.Wrap(err, "list cluster metadatas failed")
	}
	metas = p.filterObjectMetaByClusterUID(p.clusterUID, metas)
	if len(metas) == 0 {
		return DaoMetricTypes.ClusterMetricMap{}, nil
	}
	metricMap, err := p.getClusterMetricMapByObjectMetas(ctx, metas, options...)
	if err != nil {
		return DaoMetricTypes.ClusterMetricMap{}, err
	}
	metricMap.SortByTimestamp(req.QueryCondition.TimestampOrder)
	metricMap.Limit(req.QueryCondition.Limit)
	return metricMap, nil
}

func (p *ClusterMetrics) listClusterMetasFromRequest(ctx context.Context, req DaoMetricTypes.ListClusterMetricsRequest) ([]metadata.ObjectMeta, error) {

	// Generate list resource cluster request
	listClustersReq := DaoClusterStatusTypes.NewListClustersRequest()
	for index := range req.ObjectMetas {
		listClustersReq.ObjectMeta = append(listClustersReq.ObjectMeta, &req.ObjectMetas[index])
	}

	clusters, err := p.clusterStatusDAO.ListClusters(listClustersReq)
	if err != nil {
		return nil, errors.Wrap(err, "list cluster metadatas failed")
	}
	metas := make([]metadata.ObjectMeta, len(clusters))
	for i, cluster := range clusters {
		metas[i] = *cluster.ObjectMeta
	}
	return metas, nil
}

func (p *ClusterMetrics) getClusterMetricMapByObjectMetas(ctx context.Context, clusterObjectMetas []metadata.ObjectMeta, options ...DBCommon.Option) (DaoMetricTypes.ClusterMetricMap, error) {

	scope.Debugf("getClusterMetricMapByObjectMetas: clusterObjectMetas: %+v", clusterObjectMetas)

	metricMap := DaoMetricTypes.NewClusterMetricMap()
	metricChan := make(chan DaoMetricTypes.ClusterMetric)
	producerWG := errgroup.Group{}
	for _, clusterObjectMeta := range clusterObjectMetas {
		copyMeta := clusterObjectMeta
		producerWG.Go(func() error {
			m, err := p.getClusterMetric(ctx, copyMeta, options...)
			if err != nil {
				return errors.Wrap(err, "get cluster metric failed")
			}
			metricChan <- m
			return nil
		})
	}
	consumerWG := errgroup.Group{}
	consumerWG.Go(func() error {
		for m := range metricChan {
			copyM := m
			metricMap.AddClusterMetric(&copyM)
		}
		return nil
	})

	err := producerWG.Wait()
	close(metricChan)
	if err != nil {
		return metricMap, err
	}

	consumerWG.Wait()

	return metricMap, nil

}

func (p *ClusterMetrics) getClusterMetric(ctx context.Context, clusterObjectMeta metadata.ObjectMeta, options ...DBCommon.Option) (DaoMetricTypes.ClusterMetric, error) {

	emptyClusterMetric := DaoMetricTypes.ClusterMetric{
		ObjectMeta: clusterObjectMeta,
	}

	nodeMetadatas, err := p.listNodeMetasByClusterObjectMeta(ctx, clusterObjectMeta)
	if err != nil {
		return DaoMetricTypes.ClusterMetric{}, errors.Wrap(err, "list node metadatas by cluster objectmeta failed")
	} else if len(nodeMetadatas) == 0 {
		return emptyClusterMetric, nil
	}
	nodeNames := make([]string, len(nodeMetadatas))
	for i, nodeMetadata := range nodeMetadatas {
		nodeNames[i] = nodeMetadata.Name
	}

	metricChan := make(chan DaoMetricTypes.ClusterMetric)
	producerWG := errgroup.Group{}
	producerWG.Go(func() error {
		nodeCPUUsageRepo := RepoPromthMetric.NewNodeCPUUsageRepositoryWithConfig(p.PrometheusConfig)
		nodeCPUUsageEntities, err := nodeCPUUsageRepo.ListSumOfNodeCPUUsageMillicoresByNodeNames(ctx, nodeNames, options...)
		if err != nil {
			return errors.Wrap(err, "list cluster cpu usage metrics failed")
		}
		for _, e := range nodeCPUUsageEntities {
			clusterEntity := EntityPromthMetric.ClusterCPUUsageMillicoresEntity{
				ClusterName: "",
				Samples:     e.Samples,
			}
			m := clusterEntity.ClusterMetric()
			m.ObjectMeta = clusterObjectMeta
			metricChan <- m
		}
		return nil
	})
	producerWG.Go(func() error {
		nodeMemoryUsageRepo := RepoPromthMetric.NewNodeMemoryUsageRepositoryWithConfig(p.PrometheusConfig)
		nodeMemoryUsageEntities, err := nodeMemoryUsageRepo.ListSumOfNodeMetricsByNodeNames(ctx, nodeNames, options...)
		if err != nil {
			return errors.Wrap(err, "list cluster memory usage metrics failed")
		}
		for _, e := range nodeMemoryUsageEntities {
			clusterEntity := EntityPromthMetric.ClusterMemoryUsageBytesEntity{
				ClusterName: "",
				Samples:     e.Samples,
			}
			m := clusterEntity.ClusterMetric()
			m.ObjectMeta = clusterObjectMeta
			metricChan <- m
		}
		return nil
	})

	metricMap := DaoMetricTypes.NewClusterMetricMap()
	consumerWG := errgroup.Group{}
	consumerWG.Go(func() error {
		for m := range metricChan {
			copyM := m
			metricMap.AddClusterMetric(&copyM)
		}
		return nil
	})

	err = producerWG.Wait()
	close(metricChan)
	if err != nil {
		return DaoMetricTypes.ClusterMetric{}, err
	}

	consumerWG.Wait()
	metric, exist := metricMap.MetricMap[clusterObjectMeta]
	if !exist || metric == nil {
		return emptyClusterMetric, nil
	}
	return *metric, nil
}

func (p *ClusterMetrics) listNodeMetasByClusterObjectMeta(ctx context.Context, clusterObjectMeta metadata.ObjectMeta) ([]metadata.ObjectMeta, error) {

	// Generate list resource nodes request
	listNodesReq := DaoClusterStatusTypes.NewListNodesRequest()
	objectMeta := &metadata.ObjectMeta{ClusterName: clusterObjectMeta.Name}
	listNodesReq.ObjectMeta = append(listNodesReq.ObjectMeta, objectMeta)

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

func (p *ClusterMetrics) filterObjectMetaByClusterUID(clusterUID string, objectMetas []metadata.ObjectMeta) []metadata.ObjectMeta {
	newObjectMetas := make([]metadata.ObjectMeta, 0, len(objectMetas))
	for _, objectMeta := range objectMetas {
		if objectMeta.Name == clusterUID {
			newObjectMetas = append(newObjectMetas, objectMeta)
		}
	}
	return newObjectMetas
}
