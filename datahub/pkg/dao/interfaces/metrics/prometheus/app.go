package prometheus

import (
	"context"

	EntityPromthMetric "github.com/containers-ai/alameda/datahub/pkg/dao/entities/prometheus/metrics"
	DaoClusterStatusTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	RepoPromthMetric "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/prometheus/metrics"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type AppMetrics struct {
	PrometheusConfig InternalPromth.Config

	appDAO DaoClusterStatusTypes.ApplicationDAO

	clusterUID string
}

// NewAppMetricsWithConfig Constructor of prometheus app metric dao
func NewAppMetricsWithConfig(config InternalPromth.Config, appDAO DaoClusterStatusTypes.ApplicationDAO, clusterUID string) DaoMetricTypes.AppMetricsDAO {
	return &AppMetrics{
		PrometheusConfig: config,

		appDAO: appDAO,

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
	apps, err := p.listAppFromRequest(ctx, req)
	if err != nil {
		return DaoMetricTypes.AppMetricMap{}, errors.Wrap(err, "list app metadatas from request failed")
	}
	apps = p.filterApplicationsByClusterUID(p.clusterUID, apps)
	if len(apps) == 0 {
		return DaoMetricTypes.AppMetricMap{}, nil
	}
	metricMap, err = p.getAppMetricMapByApps(ctx, apps, options...)
	if err != nil {
		return DaoMetricTypes.AppMetricMap{}, errors.Wrap(err, "list app metrics failed")
	}
	metricMap.SortByTimestamp(req.QueryCondition.TimestampOrder)
	metricMap.Limit(req.QueryCondition.Limit)
	return metricMap, nil
}

func (p *AppMetrics) listAppFromRequest(ctx context.Context, req DaoMetricTypes.ListAppMetricsRequest) ([]DaoClusterStatusTypes.Application, error) {

	// Generate list resource application request
	listApplicationsReq := DaoClusterStatusTypes.NewListApplicationsRequest()
	for index := range req.ObjectMetas {
		applicationObjectMeta := DaoClusterStatusTypes.NewApplicationObjectMeta(&req.ObjectMetas[index], "")
		listApplicationsReq.ApplicationObjectMeta = append(listApplicationsReq.ApplicationObjectMeta, applicationObjectMeta)
	}

	apps, err := p.appDAO.ListApplications(listApplicationsReq)
	if err != nil {
		return nil, errors.Wrap(err, "list application metadatas failed")
	}
	nonPointerSlice := make([]DaoClusterStatusTypes.Application, len(apps))
	for i, app := range apps {
		nonPointerSlice[i] = *app
	}
	return nonPointerSlice, nil
}

func (p *AppMetrics) getAppMetricMapByApps(ctx context.Context, apps []DaoClusterStatusTypes.Application, options ...DBCommon.Option) (DaoMetricTypes.AppMetricMap, error) {
	scope.Debugf("getAppMetricMapByApps: apps: %+v", apps)

	metricMap := DaoMetricTypes.NewAppMetricMap()
	metricChan := make(chan DaoMetricTypes.AppMetric)
	producerWG := errgroup.Group{}
	for _, app := range apps {
		copyApp := app
		producerWG.Go(func() error {
			m, err := p.getAppMetric(ctx, copyApp, options...)
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

func (p *AppMetrics) getAppMetric(ctx context.Context, app DaoClusterStatusTypes.Application, options ...DBCommon.Option) (DaoMetricTypes.AppMetric, error) {

	appMeta := *app.ObjectMeta
	emptyAppMetric := DaoMetricTypes.AppMetric{
		ObjectMeta: appMeta,
	}

	controllers := p.listControllerMetasByApp(app)
	if len(controllers) == 0 {
		return emptyAppMetric, nil
	}

	namespace := controllers[0].Namespace
	podNameRegExps, err := listPodNamesRegExpByControllerObjectMetas(controllers)
	if err != nil {
		return emptyAppMetric, errors.Wrap(err, "get pod name regular expressions from controller metadata failed")
	}
	metricMap := DaoMetricTypes.NewAppMetricMap()
	metricChan := make(chan DaoMetricTypes.AppMetric)
	producerWG := errgroup.Group{}
	producerWG.Go(func() error {
		podCPUUsageRepo := RepoPromthMetric.NewPodCPUUsageRepositoryWithConfig(p.PrometheusConfig)
		podCPUMetricEntities, err := podCPUUsageRepo.ListPodCPUUsageMillicoresEntitiesBySummingPodMetrics(ctx, namespace, podNameRegExps, options...)
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
		podMemoryMetricEntities, err := podMemoryUsageRepo.ListPodMemoryUsageBytesEntityBySummingPodMetrics(ctx, namespace, podNameRegExps, options...)
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

func (p *AppMetrics) listControllerMetasByApp(app DaoClusterStatusTypes.Application) []DaoMetricTypes.ControllerObjectMeta {

	metas := make([]DaoMetricTypes.ControllerObjectMeta, len(app.Controllers))
	for i, controller := range app.Controllers {
		metas[i] = DaoMetricTypes.ControllerObjectMeta{
			ObjectMeta: *controller.ObjectMeta,
			Kind:       controller.Kind,
		}
	}

	return metas
}

func (p *AppMetrics) filterApplicationsByClusterUID(clusterUID string, apps []DaoClusterStatusTypes.Application) []DaoClusterStatusTypes.Application {
	newApps := make([]DaoClusterStatusTypes.Application, 0, len(apps))
	for _, app := range apps {
		if app.ObjectMeta.ClusterName == clusterUID {
			newApps = append(newApps, app)
		}
	}
	return newApps
}
