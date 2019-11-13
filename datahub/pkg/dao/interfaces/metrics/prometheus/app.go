package prometheus

import (
	"context"

	EntityPromthMetric "github.com/containers-ai/alameda/datahub/pkg/dao/entities/prometheus/metrics"
	DaoClusterStatusTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	RepoInfluxClusterStatus "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/clusterstatus"
	RepoPromthMetric "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/prometheus/metrics"
	"github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type AppMetrics struct {
	PrometheusConfig InternalPromth.Config
	InfluxDBConfig   InternalInflux.Config

	influxAppRepo *RepoInfluxClusterStatus.ApplicationRepository
	influxPodRepo *RepoInfluxClusterStatus.PodRepository

	clusterUID string
}

// NewAppMetricsWithConfig Constructor of prometheus app metric dao
func NewAppMetricsWithConfig(config InternalPromth.Config, influxCfg InternalInflux.Config, clusterUID string) DaoMetricTypes.AppMetricsDAO {
	return &AppMetrics{
		PrometheusConfig: config,
		InfluxDBConfig:   influxCfg,

		influxPodRepo: RepoInfluxClusterStatus.NewPodRepository(&influxCfg),
		influxAppRepo: RepoInfluxClusterStatus.NewApplicationRepositoryWithConfig(influxCfg),

		clusterUID: clusterUID,
	}
}

func (p AppMetrics) CreateMetrics(ctx context.Context, m DaoMetricTypes.AppMetricMap) error {
	return errors.New("not implemented")
}

func (p AppMetrics) ListMetrics(ctx context.Context, req DaoMetricTypes.ListAppMetricsRequest) (DaoMetricTypes.AppMetricMap, error) {

	options := []DBCommon.Option{
		DBCommon.StartTime(req.StartTime),
		DBCommon.EndTime(req.EndTime),
		DBCommon.StepTime(req.StepTime),
		DBCommon.AggregateOverTimeFunc(req.AggregateOverTimeFunction),
	}

	var metricMap DaoMetricTypes.AppMetricMap
	var err error
	appMetas, err := p.listAppMetasFromRequest(ctx, req)
	if err != nil {
		return DaoMetricTypes.AppMetricMap{}, errors.Wrap(err, "list app metadatas from request failed")
	}
	appMetas = filterObjectMetaByClusterUID(p.clusterUID, appMetas)
	if len(appMetas) == 0 {
		return DaoMetricTypes.AppMetricMap{}, nil
	}
	metricMap, err = p.getAppMetricMapByObjectMetas(ctx, appMetas, options...)
	if err != nil {
		return DaoMetricTypes.AppMetricMap{}, errors.Wrap(err, "list app metrics failed")
	}
	metricMap.SortByTimestamp(req.QueryCondition.TimestampOrder)
	metricMap.Limit(req.QueryCondition.Limit)
	return metricMap, nil
}

func (p *AppMetrics) listAppMetasFromRequest(ctx context.Context, req DaoMetricTypes.ListAppMetricsRequest) ([]metadata.ObjectMeta, error) {

	apps, err := p.influxAppRepo.ListApplications(DaoClusterStatusTypes.ListApplicationsRequest{
		ObjectMeta: req.ObjectMetas,
	})
	if err != nil {
		return nil, errors.Wrap(err, "list application metadatas failed")
	}
	metas := make([]metadata.ObjectMeta, len(apps))
	for i, app := range apps {
		metas[i] = app.ObjectMeta
	}
	return metas, nil
}

func (p *AppMetrics) getAppMetricMapByObjectMetas(ctx context.Context, appMetas []metadata.ObjectMeta, options ...DBCommon.Option) (DaoMetricTypes.AppMetricMap, error) {
	scope.Debugf("getAppMetricMapByObjectMetas: appMetas: %+v", appMetas)

	metricMap := DaoMetricTypes.NewAppMetricMap()
	metricChan := make(chan DaoMetricTypes.AppMetric)
	producerWG := errgroup.Group{}
	for _, appMeta := range appMetas {
		copyAppMeta := appMeta
		producerWG.Go(func() error {
			m, err := p.getAppMetric(ctx, copyAppMeta, options...)
			if err != nil {
				return errors.Wrap(err, "get app metric failed")
			}
			metricChan <- m
			return nil
		})
	}
	consumerWG := errgroup.Group{}
	consumerWG.Go(func() error {
		for m := range metricChan {
			copyM := m
			metricMap.AddAppMetric(&copyM)
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

func (p *AppMetrics) getAppMetric(ctx context.Context, appMeta metadata.ObjectMeta, options ...DBCommon.Option) (DaoMetricTypes.AppMetric, error) {

	emptyAppMetric := DaoMetricTypes.AppMetric{
		ObjectMeta: appMeta,
	}

	pods, err := p.listPodMetasByAppObjectMeta(ctx, appMeta)
	if err != nil {
		return emptyAppMetric, errors.Wrap(err, "list monitored pods failed")
	} else if len(pods) == 0 {
		return emptyAppMetric, nil
	}

	namespace := pods[0].Namespace
	podNames := make([]string, len(pods))
	for i, pod := range pods {
		podNames[i] = pod.Name
	}
	metricMap := DaoMetricTypes.NewAppMetricMap()
	metricChan := make(chan DaoMetricTypes.AppMetric)
	producerWG := errgroup.Group{}
	producerWG.Go(func() error {
		podCPUUsageRepo := RepoPromthMetric.NewPodCPUUsageRepositoryWithConfig(p.PrometheusConfig)
		podCPUMetricEntities, err := podCPUUsageRepo.ListPodCPUUsageMillicoresEntitiesBySummingPodMetrics(ctx, namespace, podNames, options...)
		if err != nil {
			return errors.Wrap(err, "list sum of pod cpu usage metrics failed")
		}
		for _, e := range podCPUMetricEntities {
			appEntity := EntityPromthMetric.AppCPUUsageMillicoresEntity{
				NamespaceName: appMeta.Namespace,
				AppName:       appMeta.Name,
				Samples:       e.Samples,
			}
			m := appEntity.AppMetric()
			m.ObjectMeta = appMeta
			metricChan <- m
		}
		return nil
	})
	producerWG.Go(func() error {
		podMemoryUsageRepo := RepoPromthMetric.NewPodMemoryUsageRepositoryWithConfig(p.PrometheusConfig)
		podMemoryMetricEntities, err := podMemoryUsageRepo.ListPodMemoryUsageBytesEntityBySummingPodMetrics(ctx, namespace, podNames, options...)
		if err != nil {
			return errors.Wrap(err, "list sum of pod memory usage metrics failed")
		}
		for _, e := range podMemoryMetricEntities {
			appEntity := EntityPromthMetric.AppMemoryUsageBytesEntity{
				NamespaceName: appMeta.Namespace,
				AppName:       appMeta.Name,
				Samples:       e.Samples,
			}
			m := appEntity.AppMetric()
			m.ObjectMeta = appMeta
			metricChan <- m
		}
		return nil
	})

	consumerWG := errgroup.Group{}
	consumerWG.Go(func() error {
		for m := range metricChan {
			copyM := m
			metricMap.AddAppMetric(&copyM)
		}
		return nil
	})

	err = producerWG.Wait()
	close(metricChan)
	if err != nil {
		return DaoMetricTypes.AppMetric{}, err
	}

	consumerWG.Wait()
	metric, exist := metricMap.MetricMap[appMeta]
	if !exist || metric == nil {
		return emptyAppMetric, nil
	}
	return *metric, nil
}

func (p *AppMetrics) listPodMetasByAppObjectMeta(ctx context.Context, appObjectMeta metadata.ObjectMeta) ([]metadata.ObjectMeta, error) {

	pods, err := p.influxPodRepo.ListPods(DaoClusterStatusTypes.ListPodsRequest{
		ObjectMeta: []metadata.ObjectMeta{appObjectMeta},
		Kind:       "ALAMEDASCALER",
	})
	if err != nil {
		return nil, errors.Wrap(err, "list pod metadatas by application failed")
	}
	podMetas := make([]metadata.ObjectMeta, len(pods))
	for i, pod := range pods {
		podMetas[i] = *pod.ObjectMeta
	}

	return podMetas, nil
}
