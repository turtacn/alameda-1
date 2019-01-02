package prometheus

import (
	"errors"

	"github.com/containers-ai/alameda/datahub/pkg/dao/metric"
	"github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/containerCPUUsagePercentage"
	"github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/containerMemoryUsageBytes"
	"github.com/containers-ai/alameda/datahub/pkg/repository/prometheus"
	promRepository "github.com/containers-ai/alameda/datahub/pkg/repository/prometheus/metric"
)

type prometheusMetricDAOImpl struct {
	prometheusConfig prometheus.Config
}

// NewWithConfig Constructor of prometheus metric dao
func NewWithConfig(config prometheus.Config) metric.MetricsDAO {
	return &prometheusMetricDAOImpl{prometheusConfig: config}
}

// ListPodMetrics Method implementation of MetricsDAO
func (p *prometheusMetricDAOImpl) ListPodMetrics(req metric.ListPodMetricsRequest) (metric.PodsMetricMap, error) {

	var (
		err error

		podContainerCPURepo     promRepository.PodContainerCPUUsagePercentageRepository
		podContainerMemoryRepo  promRepository.PodContainerMemoryUsageBytesRepository
		containerCPUEntities    []prometheus.Entity
		containerMemoryEntities []prometheus.Entity

		podsMetricMap    = metric.PodsMetricMap{}
		ptrPodsMetricMap = &podsMetricMap
	)

	podContainerCPURepo = promRepository.NewPodContainerCPUUsagePercentageRepositoryWithConfig(p.prometheusConfig)
	containerCPUEntities, err = podContainerCPURepo.ListMetricsByPodNamespacedName(req.Namespace, req.PodName, req.StartTime, req.EndTime)
	if err != nil {
		return podsMetricMap, errors.New("list pod metrics failed: " + err.Error())
	}

	for _, entity := range containerCPUEntities {
		containerCPUEntity := containerCPUUsagePercentage.NewEntityFromPrometheusEntity(entity)
		containerMetric := containerCPUEntity.ContainerMetric()
		ptrPodsMetricMap.AddContainerMetric(containerMetric)
	}

	podContainerMemoryRepo = promRepository.NewPodContainerMemoryUsageBytesRepositoryWithConfig(p.prometheusConfig)
	containerMemoryEntities, err = podContainerMemoryRepo.ListMetricsByPodNamespacedName(req.Namespace, req.PodName, req.StartTime, req.EndTime)
	if err != nil {
		return podsMetricMap, errors.New("list pod metrics failed: " + err.Error())
	}

	for _, entity := range containerMemoryEntities {
		containerMemoryEntity := containerMemoryUsageBytes.NewEntityFromPrometheusEntity(entity)
		containerMetric := containerMemoryEntity.ContainerMetric()
		ptrPodsMetricMap.AddContainerMetric(containerMetric)
	}

	return *ptrPodsMetricMap, nil
}
