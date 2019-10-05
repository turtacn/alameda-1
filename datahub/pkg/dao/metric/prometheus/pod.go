package prometheus

import (
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/metric/types"
	EntityPromthMetric "github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/metric"
	RepoPromthMetric "github.com/containers-ai/alameda/datahub/pkg/repository/prometheus/metric"
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

// ListMetrics Method implementation of PodMetricsDAO
func (p *PodMetrics) ListMetrics(req DaoMetricTypes.ListPodMetricsRequest) (DaoMetricTypes.PodsMetricMap, error) {

	var (
		err error

		podContainerCPURepo     RepoPromthMetric.PodContainerCpuUsagePercentageRepository
		podContainerMemoryRepo  RepoPromthMetric.PodContainerMemoryUsageBytesRepository
		containerCPUEntities    []InternalPromth.Entity
		containerMemoryEntities []InternalPromth.Entity

		podsMetricMap    = DaoMetricTypes.PodsMetricMap{}
		ptrPodsMetricMap = &podsMetricMap
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
		return podsMetricMap, errors.Wrap(err, "list pod metrics failed")
	}

	for _, entity := range containerCPUEntities {
		containerCPUEntity := EntityPromthMetric.NewContainerCpuUsagePercentageEntity(entity)
		containerMetric := containerCPUEntity.ContainerMetric()
		ptrPodsMetricMap.AddContainerMetric(&containerMetric)
	}

	podContainerMemoryRepo = RepoPromthMetric.NewPodContainerMemoryUsageBytesRepositoryWithConfig(p.PrometheusConfig)
	containerMemoryEntities, err = podContainerMemoryRepo.ListMetricsByPodNamespacedName(req.Namespace, req.PodName, options...)
	if err != nil {
		return podsMetricMap, errors.Wrap(err, "list pod metrics failed")
	}

	for _, entity := range containerMemoryEntities {
		containerMemoryEntity := EntityPromthMetric.NewContainerMemoryUsageBytesEntity(entity)
		containerMetric := containerMemoryEntity.ContainerMetric()
		ptrPodsMetricMap.AddContainerMetric(&containerMetric)
	}

	ptrPodsMetricMap.SortByTimestamp(req.QueryCondition.TimestampOrder)
	ptrPodsMetricMap.Limit(req.QueryCondition.Limit)

	return *ptrPodsMetricMap, nil
}
