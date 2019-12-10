package prometheus

import (
	"context"

	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"

	DaoClusterStatusTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	RepoPromthMetric "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/prometheus/metrics"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
)

type NamespaceMetrics struct {
	PrometheusConfig InternalPromth.Config

	namespaceDAO DaoClusterStatusTypes.NamespaceDAO

	clusterUID string
}

// NewNamespaceMetricsWithConfig Constructor of prometheus namespace metric dao
func NewNamespaceMetricsWithConfig(config InternalPromth.Config, namespaceDAO DaoClusterStatusTypes.NamespaceDAO, clusterUID string) DaoMetricTypes.NamespaceMetricsDAO {
	return &NamespaceMetrics{
		PrometheusConfig: config,

		namespaceDAO: namespaceDAO,

		clusterUID: clusterUID,
	}
}

func (p NamespaceMetrics) CreateMetrics(ctx context.Context, m DaoMetricTypes.NamespaceMetricMap) error {
	return errors.New("not implemented")
}

func (p NamespaceMetrics) ListMetrics(ctx context.Context, req DaoMetricTypes.ListNamespaceMetricsRequest) (DaoMetricTypes.NamespaceMetricMap, error) {

	options := []DBCommon.Option{
		DBCommon.StartTime(req.StartTime),
		DBCommon.EndTime(req.EndTime),
		DBCommon.StepTime(req.StepTime),
		DBCommon.AggregateOverTimeFunc(req.AggregateOverTimeFunction),
	}
	namespaceMetas, err := p.listNamespaceMetasFromRequest(ctx, req)
	if err != nil {
		return DaoMetricTypes.NamespaceMetricMap{}, errors.Wrap(err, "list namespace metadatas from request failed")
	}
	namespaceMetas = filterObjectMetaByClusterUID(p.clusterUID, namespaceMetas)
	if len(namespaceMetas) == 0 {
		return DaoMetricTypes.NamespaceMetricMap{}, nil
	}
	metricMap, err := p.getNamespaceMetricMapByObjectMetas(ctx, namespaceMetas, options...)
	if err != nil {
		return DaoMetricTypes.NamespaceMetricMap{}, err
	}
	metricMap.SortByTimestamp(req.QueryCondition.TimestampOrder)
	metricMap.Limit(req.QueryCondition.Limit)
	return metricMap, nil
}

func (p *NamespaceMetrics) listNamespaceMetasFromRequest(ctx context.Context, req DaoMetricTypes.ListNamespaceMetricsRequest) ([]metadata.ObjectMeta, error) {

	// Generate list resource namespaces request
	listNamespacesReq := DaoClusterStatusTypes.NewListNamespacesRequest()
	for index := range req.ObjectMetas {
		listNamespacesReq.ObjectMeta = append(listNamespacesReq.ObjectMeta, &req.ObjectMetas[index])
	}

	namespaces, err := p.namespaceDAO.ListNamespaces(listNamespacesReq)
	if err != nil {
		return nil, errors.Wrap(err, "list namespaces metadatas failed")
	}
	metas := make([]metadata.ObjectMeta, len(namespaces))
	for i, namespace := range namespaces {
		metas[i] = *namespace.ObjectMeta
	}
	return metas, nil
}

func (p *NamespaceMetrics) getNamespaceMetricMapByObjectMetas(ctx context.Context, namespaceMetas []metadata.ObjectMeta, options ...DBCommon.Option) (DaoMetricTypes.NamespaceMetricMap, error) {
	scope.Debugf("getNamespaceMetricMapByObjectMetas: namespaceMetas: %+v", namespaceMetas)

	// Build namespace map for later searching
	namespaceMetaMap := make(map[string]metadata.ObjectMeta)
	for _, meta := range namespaceMetas {
		namespaceMetaMap[meta.Name] = meta
	}

	namespaceNames := make([]string, len(namespaceMetas))
	for i, meta := range namespaceMetas {
		namespaceNames[i] = meta.Name
	}
	metricChan := make(chan DaoMetricTypes.NamespaceMetric)
	producerWG := errgroup.Group{}
	producerWG.Go(func() error {

		namespaceCPUUsageRepo := RepoPromthMetric.NewNamespaceCPUUsageRepositoryWithConfig(p.PrometheusConfig)
		namespaceCPUUsageEntities, err := namespaceCPUUsageRepo.ListNamespaceCPUUsageMillicoresEntitiesByNamespaceNames(ctx, namespaceNames, options...)
		if err != nil {
			return errors.Wrap(err, "list namespace cpu usage metrics failed")
		}
		for _, e := range namespaceCPUUsageEntities {
			m := e.NamespaceMetric()
			m.ObjectMeta = namespaceMetaMap[e.NamespaceName]
			metricChan <- m
		}

		return nil
	})
	producerWG.Go(func() error {

		namespaceMemoryUsageRepo := RepoPromthMetric.NewNamespaceMemoryUsageRepositoryWithConfig(p.PrometheusConfig)
		namespaceMemoryUsageEntities, err := namespaceMemoryUsageRepo.ListNamespaceMemoryUsageBytesEntitiesByNamespaceNames(ctx, namespaceNames, options...)
		if err != nil {
			return errors.Wrap(err, "list namespace memory usage metrics failed")
		}
		for _, e := range namespaceMemoryUsageEntities {
			m := e.NamespaceMetric()
			m.ObjectMeta = namespaceMetaMap[e.NamespaceName]
			metricChan <- m
		}

		return nil
	})

	metricMap := DaoMetricTypes.NewNamespaceMetricMap()
	consumerWG := errgroup.Group{}
	consumerWG.Go(func() error {

		for m := range metricChan {
			copyM := m
			metricMap.AddNamespaceMetric(&copyM)
		}

		return nil
	})

	err := producerWG.Wait()
	close(metricChan)
	if err != nil {
		return metricMap, err
	}

	consumerWG.Wait()
	for _, namespaceMeta := range namespaceMetas {
		if metric, exist := metricMap.MetricMap[namespaceMeta]; !exist || metric == nil {
			metricMap.MetricMap[namespaceMeta] = &DaoMetricTypes.NamespaceMetric{
				ObjectMeta: namespaceMeta,
			}
		}
	}

	return metricMap, nil
}
