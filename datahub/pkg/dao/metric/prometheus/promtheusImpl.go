package prometheus

import (
	"github.com/containers-ai/alameda/datahub/pkg/dao/metric"
	EntityPromthMetric "github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/metric"
	RepoPromthMetric "github.com/containers-ai/alameda/datahub/pkg/repository/prometheus/metric"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/pkg/errors"
	"sync"
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

		podContainerCPURepo     RepoPromthMetric.PodContainerCpuUsagePercentageRepository
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

	podContainerCPURepo = RepoPromthMetric.NewPodContainerCpuUsagePercentageRepositoryWithConfig(p.prometheusConfig)
	containerCPUEntities, err = podContainerCPURepo.ListMetricsByPodNamespacedName(req.Namespace, req.PodName, options...)
	if err != nil {
		return podsMetricMap, errors.Wrap(err, "list pod metrics failed")
	}

	for _, entity := range containerCPUEntities {
		containerCPUEntity := EntityPromthMetric.NewContainerCpuUsagePercentageEntity(entity)
		containerMetric := containerCPUEntity.ContainerMetric()
		ptrPodsMetricMap.AddContainerMetric(&containerMetric)
	}

	podContainerMemoryRepo = RepoPromthMetric.NewPodContainerMemoryUsageBytesRepositoryWithConfig(p.prometheusConfig)
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

// ListNodesMetric Method implementation of MetricsDAO
func (p *prometheusMetricDAOImpl) ListNodesMetric(req metric.ListNodeMetricsRequest) (metric.NodesMetricMap, error) {

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
		return metric.NodesMetricMap{}, errors.Wrap(err, "list nodes metrics failed")
	}

	ptrNodesMetricMap.SortByTimestamp(req.QueryCondition.TimestampOrder)
	ptrNodesMetricMap.Limit(req.QueryCondition.Limit)

	return *ptrNodesMetricMap, nil
}

func (p *prometheusMetricDAOImpl) produceNodeMetric(nodeName string, nodeMetricChan chan<- metric.NodeMetric, errChan chan<- error, wg *sync.WaitGroup, options ...DBCommon.Option) {

	var (
		err                     error
		nodeCPUUsageRepo        RepoPromthMetric.NodeCpuUsagePercentageRepository
		nodeMemoryUsageRepo     RepoPromthMetric.NodeMemoryBytesUsageRepository
		nodeCPUUsageEntities    []InternalPromth.Entity
		nodeMemoryUsageEntities []InternalPromth.Entity
	)

	defer wg.Done()

	nodeCPUUsageRepo = RepoPromthMetric.NewNodeCpuUsagePercentageRepositoryWithConfig(p.prometheusConfig)
	nodeCPUUsageEntities, err = nodeCPUUsageRepo.ListMetricsByNodeName(nodeName, options...)
	if err != nil {
		errChan <- errors.Wrap(err, "list node cpu usage metrics failed")
		return
	}

	for _, entity := range nodeCPUUsageEntities {
		nodeCPUUsageEntity := EntityPromthMetric.NewNodeCpuUsagePercentageEntity(entity)
		nodeMetric := nodeCPUUsageEntity.NodeMetric()
		nodeMetricChan <- nodeMetric
	}

	nodeMemoryUsageRepo = RepoPromthMetric.NewNodeMemoryBytesUsageRepositoryWithConfig(p.prometheusConfig)
	nodeMemoryUsageEntities, err = nodeMemoryUsageRepo.ListMetricsByNodeName(nodeName, options...)
	if err != nil {
		errChan <- errors.Wrap(err, "list node memory usage metrics failed")
		return
	}

	for _, entity := range nodeMemoryUsageEntities {
		noodeMemoryUsageEntity := EntityPromthMetric.NewNodeMemoryBytesUsageEntity(entity)
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
