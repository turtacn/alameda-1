package prometheus

import (
	EntityPromthMetric "github.com/containers-ai/alameda/datahub/pkg/dao/entities/prometheus/metrics"
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/metrics/types"
	RepoPromthMetric "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/prometheus/metrics"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/pkg/errors"
)

type PodMetrics struct {
	PrometheusConfig InternalPromth.Config
}

// NewPodMetricsWithConfig Constructor of prometheus pod metric dao
func NewPodMetricsWithConfig(config InternalPromth.Config) DaoMetricTypes.PodMetricsDAO {
	return &PodMetrics{PrometheusConfig: config}
}

// CreateMetrics Method implementation of PodMetricsDAO
func (p *PodMetrics) CreateMetrics(metrics DaoMetricTypes.PodMetricMap) error {
	return errors.New("create pod metrics to prometheus is not supported")
}

// ListMetrics Method implementation of PodMetricsDAO
func (p *PodMetrics) ListMetrics(req DaoMetricTypes.ListPodMetricsRequest) (DaoMetricTypes.PodMetricMap, error) {
	var (
		err error

		podContainerCPURepo     RepoPromthMetric.PodContainerCpuUsagePercentageRepository
		podContainerMemoryRepo  RepoPromthMetric.PodContainerMemoryUsageBytesRepository
		containerCPUEntities    []InternalPromth.Entity
		containerMemoryEntities []InternalPromth.Entity

		podMetricMap    = DaoMetricTypes.NewPodMetricMap()
		ptrPodMetricMap = &podMetricMap
	)

	options := []DBCommon.Option{
		DBCommon.StartTime(req.StartTime),
		DBCommon.EndTime(req.EndTime),
		DBCommon.StepTime(req.StepTime),
		DBCommon.AggregateOverTimeFunc(req.AggregateOverTimeFunction),
	}

	podContainerCPURepo = RepoPromthMetric.NewPodContainerCpuUsagePercentageRepositoryWithConfig(p.PrometheusConfig)
	containerCPUEntities, err = podContainerCPURepo.ListMetricsByPodNamespacedName(req.Namespace, req.PodName, options...)
	if err != nil {
		return podMetricMap, errors.Wrap(err, "list pod metrics failed")
	}

	for _, entity := range containerCPUEntities {
		containerCPUEntity := EntityPromthMetric.NewContainerCpuUsagePercentageEntity(entity)
		containerMetric := containerCPUEntity.ContainerMetric()
		ptrPodMetricMap.AddContainerMetric(&containerMetric)
	}

	podContainerMemoryRepo = RepoPromthMetric.NewPodContainerMemoryUsageBytesRepositoryWithConfig(p.PrometheusConfig)
	containerMemoryEntities, err = podContainerMemoryRepo.ListMetricsByPodNamespacedName(req.Namespace, req.PodName, options...)
	if err != nil {
		return podMetricMap, errors.Wrap(err, "list pod metrics failed")
	}

	for _, entity := range containerMemoryEntities {
		containerMemoryEntity := EntityPromthMetric.NewContainerMemoryUsageBytesEntity(entity)
		containerMetric := containerMemoryEntity.ContainerMetric()
		ptrPodMetricMap.AddContainerMetric(&containerMetric)
	}

	ptrPodMetricMap.SortByTimestamp(req.QueryCondition.TimestampOrder)
	ptrPodMetricMap.Limit(req.QueryCondition.Limit)

	return *ptrPodMetricMap, nil
}
