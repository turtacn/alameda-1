package prometheus

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/metric"
	"github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/containerCPUUsagePercentage"
	"github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/containerMemoryUsageBytes"
	"github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/nodeCPUUsagePercentage"
	"github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/nodeMemoryUsageBytes"
	RepoPromthMetric "github.com/containers-ai/alameda/datahub/pkg/repository/prometheus/metric"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/containers-ai/alameda/pkg/utils/log"
	"github.com/pkg/errors"
	"sync"
)

var (
	scope = log.RegisterScope("metric_dao_implement", "metric dao implement", 0)
)

type prometheusMetricDAOImpl struct {
	prometheusConfig InternalPromth.Config
}

// NewWithConfig Constructor of prometheus metric dao
func NewWithConfig(config InternalPromth.Config) metric.MetricsDAO {
	return &prometheusMetricDAOImpl{prometheusConfig: config}
}

// ListPodMetrics Method implementation of MetricsDAO
func (p *prometheusMetricDAOImpl) ListPodMetrics(req metric.ListPodMetricsRequest) (metric.PodsMetricMap, error) {

	var (
		err error

		podContainerCPURepo     RepoPromthMetric.PodContainerCPUUsagePercentageRepository
		podContainerMemoryRepo  RepoPromthMetric.PodContainerMemoryUsageBytesRepository
		containerCPUEntities    []InternalPromth.Entity
		containerMemoryEntities []InternalPromth.Entity

		podsMetricMap    = metric.PodsMetricMap{}
		ptrPodsMetricMap = &podsMetricMap
	)

	options := []DBCommon.Option{
		DBCommon.StartTime(req.StartTime),
		DBCommon.EndTime(req.EndTime),
		DBCommon.StepTime(req.StepTime),
		DBCommon.AggregateOverTimeFunc(req.AggregateOverTimeFunction),
	}

	podContainerCPURepo = RepoPromthMetric.NewPodContainerCPUUsagePercentageRepositoryWithConfig(p.prometheusConfig)
	containerCPUEntities, err = podContainerCPURepo.ListMetricsByPodNamespacedName(req.Namespace, req.PodName, options...)
	if err != nil {
		return podsMetricMap, errors.Wrap(err, "list pod metrics failed")
	}

	for _, entity := range containerCPUEntities {
		containerCPUEntity := containerCPUUsagePercentage.NewEntityFromPrometheusEntity(entity)
		containerMetric := containerCPUEntity.ContainerMetric()
		ptrPodsMetricMap.AddContainerMetric(&containerMetric)
	}

	podContainerMemoryRepo = RepoPromthMetric.NewPodContainerMemoryUsageBytesRepositoryWithConfig(p.prometheusConfig)
	containerMemoryEntities, err = podContainerMemoryRepo.ListMetricsByPodNamespacedName(req.Namespace, req.PodName, options...)
	if err != nil {
		return podsMetricMap, errors.Wrap(err, "list pod metrics failed")
	}

	for _, entity := range containerMemoryEntities {
		containerMemoryEntity := containerMemoryUsageBytes.NewEntityFromPrometheusEntity(entity)
		containerMetric := containerMemoryEntity.ContainerMetric()
		ptrPodsMetricMap.AddContainerMetric(&containerMetric)
	}

	ptrPodsMetricMap.SortByTimestamp(req.QueryCondition.TimestampOrder)
	ptrPodsMetricMap.Limit(req.QueryCondition.Limit)

	return *ptrPodsMetricMap, nil
}

// ListNodesMetric Method implementation of MetricsDAO
func (p *prometheusMetricDAOImpl) ListNodesMetric(req metric.ListNodeMetricsRequest) (metric.NodesMetricMap, error) {

	scope.Infof("dao-metrics-ListNodesMetric input %v", &req)
	var (
		wg             = sync.WaitGroup{}
		nodeNames      []string
		nodeMetricChan = make(chan metric.NodeMetric)
		errChan        chan error
		done           = make(chan bool)

		nodesMetricMap    = metric.NodesMetricMap{}
		ptrNodesMetricMap = &nodesMetricMap
	)

	if len(req.GetNodeNames()) != 0 {
		nodeNames = req.GetNodeNames()
	} else {
		nodeNames = req.GetEmptyNodeNames()
	}

	options := []DBCommon.Option{
		DBCommon.StartTime(req.StartTime),
		DBCommon.EndTime(req.EndTime),
		DBCommon.StepTime(req.StepTime),
		DBCommon.AggregateOverTimeFunc(req.AggregateOverTimeFunction),
	}

	errChan = make(chan error, len(nodeNames))
	wg.Add(len(nodeNames))
	for _, nodeName := range nodeNames {
		go p.produceNodeMetric(nodeName, nodeMetricChan, errChan, &wg, options...)
	}

	go addNodeMetricToNodesMetricMap(ptrNodesMetricMap, nodeMetricChan, done)

	wg.Wait()
	close(nodeMetricChan)

	select {
	case _ = <-done:
	case err := <-errChan:
		scope.Infof("dao-metric-ListNodesMetric error %v", err)
		return metric.NodesMetricMap{}, errors.Wrap(err, "list nodes metrics failed")
	}

	ptrNodesMetricMap.SortByTimestamp(req.QueryCondition.TimestampOrder)
	ptrNodesMetricMap.Limit(req.QueryCondition.Limit)

	scope.Infof("dao-metric-ListNodesMetric return %d %v", len(*ptrNodesMetricMap), ptrNodesMetricMap)
	return *ptrNodesMetricMap, nil
}

func (p *prometheusMetricDAOImpl) produceNodeMetric(nodeName string, nodeMetricChan chan<- metric.NodeMetric, errChan chan<- error, wg *sync.WaitGroup, options ...DBCommon.Option) {

	var (
		err                     error
		nodeCPUUsageRepo        RepoPromthMetric.NodeCPUUsagePercentageRepository
		nodeMemoryUsageRepo     RepoPromthMetric.NodeMemoryUsageBytesRepository
		nodeCPUUsageEntities    []InternalPromth.Entity
		nodeMemoryUsageEntities []InternalPromth.Entity
	)

	defer wg.Done()

	nodeCPUUsageRepo = RepoPromthMetric.NewNodeCPUUsagePercentageRepositoryWithConfig(p.prometheusConfig)
	nodeCPUUsageEntities, err = nodeCPUUsageRepo.ListMetricsByNodeName(nodeName, options...)
	if err != nil {
		errChan <- errors.Wrap(err, "list node cpu usage metrics failed")
		return
	}

	for _, entity := range nodeCPUUsageEntities {
		nodeCPUUsageEntity := nodeCPUUsagePercentage.NewEntityFromPrometheusEntity(entity)
		nodeMetric := nodeCPUUsageEntity.NodeMetric()
		nodeMetricChan <- nodeMetric
	}

	nodeMemoryUsageRepo = RepoPromthMetric.NewNodeMemoryUsageBytesRepositoryWithConfig(p.prometheusConfig)
	nodeMemoryUsageEntities, err = nodeMemoryUsageRepo.ListMetricsByNodeName(nodeName, options...)
	if err != nil {
		errChan <- errors.Wrap(err, "list node memory usage metrics failed")
		return
	}

	for _, entity := range nodeMemoryUsageEntities {
		noodeMemoryUsageEntity := nodeMemoryUsageBytes.NewEntityFromPrometheusEntity(entity)
		nodeMetric := noodeMemoryUsageEntity.NodeMetric()
		nodeMetricChan <- nodeMetric
	}
}

func addNodeMetricToNodesMetricMap(nodesMetricMap *metric.NodesMetricMap, nodeMetricChan <-chan metric.NodeMetric, done chan<- bool) {

	for {
		nodeMetric, more := <-nodeMetricChan
		if more {
			nodesMetricMap.AddNodeMetric(&nodeMetric)
		} else {
			done <- true
			return
		}
	}
}
