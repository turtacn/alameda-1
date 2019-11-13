package prometheus

import (
	"context"
	"golang.org/x/sync/errgroup"

	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	RepoPromthMetric "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/prometheus/metrics"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/pkg/errors"
)

type PodMetrics struct {
	PrometheusConfig InternalPromth.Config

	clusterUID string
}

// NewPodMetricsWithConfig Constructor of prometheus pod metric dao
func NewPodMetricsWithConfig(config InternalPromth.Config, clusterUID string) DaoMetricTypes.PodMetricsDAO {
	return &PodMetrics{
		PrometheusConfig: config,

		clusterUID: clusterUID,
	}
}

// CreateMetrics Method implementation of PodMetricsDAO
func (p *PodMetrics) CreateMetrics(ctx context.Context, metrics DaoMetricTypes.PodMetricMap) error {
	return errors.New("create pod metrics to prometheus is not supported")
}

// ListMetrics Method implementation of PodMetricsDAO
func (p *PodMetrics) ListMetrics(ctx context.Context, req DaoMetricTypes.ListPodMetricsRequest) (DaoMetricTypes.PodMetricMap, error) {

	options := []DBCommon.Option{
		DBCommon.StartTime(req.StartTime),
		DBCommon.EndTime(req.EndTime),
		DBCommon.StepTime(req.StepTime),
		DBCommon.AggregateOverTimeFunc(req.AggregateOverTimeFunction),
	}

	podMetas, err := p.listPodMetasFromRequest(ctx, req)
	if err != nil {
		return DaoMetricTypes.PodMetricMap{}, errors.Wrap(err, "list pod metadatas from request failed")
	}
	podMetas = filterObjectMetaByClusterUID(p.clusterUID, podMetas)
	if len(podMetas) == 0 {
		return DaoMetricTypes.PodMetricMap{}, nil
	}
	metricMap, err := p.getPodMetricMapByObjectMetas(ctx, podMetas, options...)
	if err != nil {
		return DaoMetricTypes.PodMetricMap{}, errors.Wrap(err, "get pod metricMap failed")
	}
	metricMap.SortByTimestamp(req.QueryCondition.TimestampOrder)
	metricMap.Limit(req.QueryCondition.Limit)
	return metricMap, nil
}

func (p *PodMetrics) listPodMetasFromRequest(ctx context.Context, req DaoMetricTypes.ListPodMetricsRequest) ([]metadata.ObjectMeta, error) {
	// TODO
	return nil, nil
}

func (p *PodMetrics) getPodMetricMapByObjectMetas(ctx context.Context, podMetas []metadata.ObjectMeta, options ...DBCommon.Option) (DaoMetricTypes.PodMetricMap, error) {

	// To minimize the times to query prometheus, aggregate pods in the same namespaces
	// map[pod.Namespace]map[pod.Name]pod.ObjectMeta
	namespacePodMap := make(map[string]map[string]metadata.ObjectMeta)
	for _, podMeta := range podMetas {
		if namespacePodMap[podMeta.Namespace] == nil {
			namespacePodMap[podMeta.Namespace] = make(map[string]metadata.ObjectMeta)
		}
		namespacePodMap[podMeta.Namespace][podMeta.Name] = podMeta
	}

	metricChan := make(chan DaoMetricTypes.ContainerMetric)
	producerWG := errgroup.Group{}
	// Query cpu metrics
	producerWG.Go(func() error {
		wg := errgroup.Group{}
		podContainerCPURepo := RepoPromthMetric.NewContainerCpuUsageRepositoryWithConfig(p.PrometheusConfig)
		for namespace, podMetaMap := range namespacePodMap {
			copyPodMetaMap := podMetaMap
			wg.Go(func() error {
				podNames := make([]string, 0, len(copyPodMetaMap))
				for podName := range copyPodMetaMap {
					podNames = append(podNames, podName)
				}
				containerCPUEntities, err := podContainerCPURepo.ListContainerCPUUsageMillicoresEntitiesByNamespaceAndPodNames(ctx, namespace, podNames, options...)
				if err != nil {
					return errors.Wrap(err, "list pod cpu usage metrics failed")
				}
				for _, e := range containerCPUEntities {
					m := e.ContainerMetric()
					clusterName := ""
					if meta, exist := copyPodMetaMap[m.ObjectMeta.PodName]; exist {
						clusterName = meta.ClusterName
					}
					m.ObjectMeta.ObjectMeta.ClusterName = clusterName
					metricChan <- m
				}
				return nil
			})
		}

		return wg.Wait()
	})
	// Query memory metrics
	producerWG.Go(func() error {
		wg := errgroup.Group{}
		podContainerMemoryRepo := RepoPromthMetric.NewContainerMemoryUsageRepositoryWithConfig(p.PrometheusConfig)
		for namespace, podMetaMap := range namespacePodMap {
			copyPodMetaMap := podMetaMap
			wg.Go(func() error {
				podNames := make([]string, 0, len(copyPodMetaMap))
				for podName := range copyPodMetaMap {
					podNames = append(podNames, podName)
				}
				containerMemoryEntities, err := podContainerMemoryRepo.ListContainerMemoryUsageBytesEntitiesByNamespaceAndPodNames(ctx, namespace, podNames, options...)
				if err != nil {
					return errors.Wrap(err, "list pod memory usage metrics failed")
				}
				for _, e := range containerMemoryEntities {
					m := e.ContainerMetric()
					clusterName := ""
					if meta, exist := copyPodMetaMap[m.ObjectMeta.PodName]; exist {
						clusterName = meta.ClusterName
					}
					m.ObjectMeta.ObjectMeta.ClusterName = clusterName
					metricChan <- m
				}
				return nil
			})
		}
		return wg.Wait()
	})

	metricMap := DaoMetricTypes.NewPodMetricMap()
	consumerWG := errgroup.Group{}
	consumerWG.Go(func() error {
		for m := range metricChan {
			copyM := m
			metricMap.AddContainerMetric(&copyM)
		}
		return nil
	})

	err := producerWG.Wait()
	close(metricChan)
	if err != nil {
		return DaoMetricTypes.PodMetricMap{}, err
	}

	consumerWG.Wait()
	for _, podMeta := range podMetas {
		if metric, exist := metricMap.MetricMap[podMeta]; !exist || metric == nil {
			metricMap.MetricMap[podMeta] = &DaoMetricTypes.PodMetric{
				ObjectMeta: podMeta,
			}
		}
	}
	return metricMap, nil
}
