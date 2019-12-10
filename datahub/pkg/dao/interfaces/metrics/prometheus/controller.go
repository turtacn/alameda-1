package prometheus

import (
	"context"
	"strings"

	EntityPromthMetric "github.com/containers-ai/alameda/datahub/pkg/dao/entities/prometheus/metrics"
	DaoClusterStatusTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	RepoPromthMetric "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/prometheus/metrics"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

type ControllerMetrics struct {
	PrometheusConfig InternalPromth.Config

	controllerDAO DaoClusterStatusTypes.ControllerDAO

	clusterUID string
}

// NewControllerMetricsWithConfig Constructor of prometheus controller metric dao
func NewControllerMetricsWithConfig(promCfg InternalPromth.Config, controllerDAO DaoClusterStatusTypes.ControllerDAO, clusterUID string) DaoMetricTypes.ControllerMetricsDAO {
	return &ControllerMetrics{
		PrometheusConfig: promCfg,

		controllerDAO: controllerDAO,

		clusterUID: clusterUID,
	}
}

func (p ControllerMetrics) CreateMetrics(ctx context.Context, m DaoMetricTypes.ControllerMetricMap) error {
	return errors.New("not implemented")
}

// ListMetrics returns controller metrics, if length of req.ControllerObjectMetas equals 0, fetch all kinds of controller metrics in the cluster
// otherwise listing controller metadatas by function listControllerMetasFromRequest and fetching the returned controllers' metrics.
func (p ControllerMetrics) ListMetrics(ctx context.Context, req DaoMetricTypes.ListControllerMetricsRequest) (DaoMetricTypes.ControllerMetricMap, error) {

	options := []DBCommon.Option{
		DBCommon.StartTime(req.StartTime),
		DBCommon.EndTime(req.EndTime),
		DBCommon.StepTime(req.StepTime),
		DBCommon.AggregateOverTimeFunc(req.AggregateOverTimeFunction),
	}

	var metricMap DaoMetricTypes.ControllerMetricMap
	var err error
	controllerMetas, err := p.listControllerMetasFromRequest(ctx, req)
	if err != nil {
		return DaoMetricTypes.ControllerMetricMap{}, errors.Wrap(err, "list controller metadatas from request failed")
	}
	controllerMetas = p.filterObjectMetaByClusterUID(p.clusterUID, controllerMetas)
	if len(controllerMetas) == 0 {
		return DaoMetricTypes.ControllerMetricMap{}, nil
	}
	metricMap, err = p.getControllerMetricMapByObjectMetas(ctx, controllerMetas, options...)
	if err != nil {
		return DaoMetricTypes.ControllerMetricMap{}, errors.Wrap(err, "list controller metrics failed")
	}
	metricMap.SortByTimestamp(req.QueryCondition.TimestampOrder)
	metricMap.Limit(req.QueryCondition.Limit)
	return metricMap, nil
}

func (p *ControllerMetrics) listControllerMetasFromRequest(ctx context.Context, req DaoMetricTypes.ListControllerMetricsRequest) ([]DaoMetricTypes.ControllerObjectMeta, error) {

	// Generate list resource controllers request
	listControllersReq := DaoClusterStatusTypes.NewListControllersRequest()
	for _, objectMeta := range req.ObjectMetas {
		controllerObjectMeta := DaoClusterStatusTypes.NewControllerObjectMeta(&objectMeta, nil, strings.ToUpper(req.Kind), "")
		listControllersReq.ControllerObjectMeta = append(listControllersReq.ControllerObjectMeta, controllerObjectMeta)

	}
	if len(listControllersReq.ControllerObjectMeta) == 0 {
		controllerObjectMeta := DaoClusterStatusTypes.NewControllerObjectMeta(nil, nil, strings.ToUpper(req.Kind), "")
		listControllersReq.ControllerObjectMeta = append(listControllersReq.ControllerObjectMeta, controllerObjectMeta)
	}

	controllers, err := p.controllerDAO.ListControllers(listControllersReq)
	if err != nil {
		return nil, errors.Wrap(err, "list controller metadatas failed")
	}
	metas := make([]DaoMetricTypes.ControllerObjectMeta, len(controllers))
	for i, controller := range controllers {
		metas[i] = DaoMetricTypes.ControllerObjectMeta{
			ObjectMeta: *controller.ObjectMeta,
			Kind:       controller.Kind,
		}
	}
	return metas, nil
}

func (p *ControllerMetrics) getControllerMetricMapByObjectMetas(ctx context.Context, controllerMetas []DaoMetricTypes.ControllerObjectMeta, options ...DBCommon.Option) (DaoMetricTypes.ControllerMetricMap, error) {
	scope.Debugf("getControllerMetricMapByObjectMetas: controllerMetas: %+v", controllerMetas)

	metricMap := DaoMetricTypes.NewControllerMetricMap()
	metricChan := make(chan DaoMetricTypes.ControllerMetric)
	producerWG := errgroup.Group{}
	for _, controllerMeta := range controllerMetas {
		copyControllerMeta := controllerMeta
		producerWG.Go(func() error {
			m, err := p.getControllerMetric(ctx, copyControllerMeta, options...)
			if err != nil {
				return errors.Wrap(err, "get controller metric failed")
			}
			metricChan <- m
			return nil
		})
	}
	consumerWG := errgroup.Group{}
	consumerWG.Go(func() error {
		for m := range metricChan {
			copyM := m
			metricMap.AddControllerMetric(&copyM)
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

func (p *ControllerMetrics) getControllerMetric(ctx context.Context, controllerMeta DaoMetricTypes.ControllerObjectMeta, options ...DBCommon.Option) (DaoMetricTypes.ControllerMetric, error) {

	emptyControllerMetric := DaoMetricTypes.ControllerMetric{
		ObjectMeta: controllerMeta,
	}

	namespace := controllerMeta.Namespace
	podNameRegExps, err := listPodNamesRegExpByControllerObjectMetas([]DaoMetricTypes.ControllerObjectMeta{controllerMeta})
	if err != nil {
		return emptyControllerMetric, errors.Wrap(err, "get pod name regular expressions from controller metadata failed")
	}

	metricMap := DaoMetricTypes.NewControllerMetricMap()
	metricChan := make(chan DaoMetricTypes.ControllerMetric)
	producerWG := errgroup.Group{}
	producerWG.Go(func() error {
		podCPUUsageRepo := RepoPromthMetric.NewPodCPUUsageRepositoryWithConfig(p.PrometheusConfig)
		podCPUMetricEntities, err := podCPUUsageRepo.ListPodCPUUsageMillicoresEntitiesBySummingPodMetrics(ctx, namespace, podNameRegExps, options...)
		if err != nil {
			return errors.Wrap(err, "list sum of pod cpu usage metrics failed")
		}
		for _, e := range podCPUMetricEntities {
			controllerEntity := EntityPromthMetric.ControllerCPUUsageMillicoresEntity{
				NamespaceName:  controllerMeta.Namespace,
				ControllerName: controllerMeta.Name,
				ControllerKind: controllerMeta.Kind,
				Samples:        e.Samples,
			}
			m := controllerEntity.ControllerMetric()
			m.ObjectMeta = controllerMeta
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
			controllerEntity := EntityPromthMetric.ControllerMemoryUsageBytesEntity{
				NamespaceName:  controllerMeta.Namespace,
				ControllerName: controllerMeta.Name,
				ControllerKind: controllerMeta.Kind,
				Samples:        e.Samples,
			}
			m := controllerEntity.ControllerMetric()
			m.ObjectMeta = controllerMeta
			metricChan <- m
		}
		return nil
	})

	consumerWG := errgroup.Group{}
	consumerWG.Go(func() error {
		for m := range metricChan {
			copyM := m
			metricMap.AddControllerMetric(&copyM)
		}
		return nil
	})

	err = producerWG.Wait()
	close(metricChan)
	if err != nil {
		return DaoMetricTypes.ControllerMetric{}, err
	}

	consumerWG.Wait()
	metric, exist := metricMap.MetricMap[controllerMeta]
	if !exist || metric == nil {
		return emptyControllerMetric, nil
	}
	return *metric, nil
}

func (p *ControllerMetrics) filterObjectMetaByClusterUID(clusterUID string, objectMetas []DaoMetricTypes.ControllerObjectMeta) []DaoMetricTypes.ControllerObjectMeta {
	newObjectMetas := make([]DaoMetricTypes.ControllerObjectMeta, 0, len(objectMetas))
	for _, objectMeta := range objectMetas {
		if objectMeta.ClusterName == clusterUID {
			newObjectMetas = append(newObjectMetas, objectMeta)
		}
	}
	return newObjectMetas
}
