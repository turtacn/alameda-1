package prometheus

import (
	DaoMetricTypes "github.com/containers-ai/alameda/datahub/pkg/dao/metric/types"
	EntityPromthMetric "github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/metric"
	RepoPromthMetric "github.com/containers-ai/alameda/datahub/pkg/repository/prometheus/metric"
	DBCommon "github.com/containers-ai/alameda/internal/pkg/database/common"
	InternalPromth "github.com/containers-ai/alameda/internal/pkg/database/prometheus"
	"github.com/pkg/errors"
	"sync"
)

type NodeMetrics struct {
	PrometheusConfig InternalPromth.Config
}

// NewNodeMetricsWithConfig Constructor of prometheus node metric dao
func NewNodeMetricsWithConfig(config InternalPromth.Config) DaoMetricTypes.NodeMetricsDAO {
	return &NodeMetrics{PrometheusConfig: config}
}

// CreateMetrics Method implementation of NodeMetricsDAO
func (p *NodeMetrics) CreateMetrics(metrics DaoMetricTypes.NodeMetricMap) error {
	return errors.New("create metrics to prometheus is not supported")
}

// ListMetrics Method implementation of NodeMetricsDAO
func (p *NodeMetrics) ListMetrics(req DaoMetricTypes.ListNodeMetricsRequest) (DaoMetricTypes.NodeMetricMap, error) {
	var (
		wg             = sync.WaitGroup{}
		nodeNames      []string
		nodeMetricChan = make(chan DaoMetricTypes.NodeMetric)
		errChan        chan error
		done           = make(chan bool)

		nodeMetricMap    = DaoMetricTypes.NewNodeMetricMap()
		ptrNodeMetricMap = &nodeMetricMap
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

	go addNodeMetricToNodesMetricMap(ptrNodeMetricMap, nodeMetricChan, done)

	wg.Wait()
	close(nodeMetricChan)

	select {
	case _ = <-done:
	case err := <-errChan:
		return DaoMetricTypes.NewNodeMetricMap(), errors.Wrap(err, "list nodes metrics failed")
	}

	ptrNodeMetricMap.SortByTimestamp(req.QueryCondition.TimestampOrder)
	ptrNodeMetricMap.Limit(req.QueryCondition.Limit)

	return *ptrNodeMetricMap, nil
}

func (p *NodeMetrics) produceNodeMetric(nodeName string, nodeMetricChan chan<- DaoMetricTypes.NodeMetric, errChan chan<- error, wg *sync.WaitGroup, options ...DBCommon.Option) {
	var (
		err                     error
		nodeCPUUsageRepo        RepoPromthMetric.NodeCpuUsagePercentageRepository
		nodeMemoryUsageRepo     RepoPromthMetric.NodeMemoryBytesUsageRepository
		nodeCPUUsageEntities    []InternalPromth.Entity
		nodeMemoryUsageEntities []InternalPromth.Entity
	)

	defer wg.Done()

	nodeCPUUsageRepo = RepoPromthMetric.NewNodeCpuUsagePercentageRepositoryWithConfig(p.PrometheusConfig)
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

	nodeMemoryUsageRepo = RepoPromthMetric.NewNodeMemoryBytesUsageRepositoryWithConfig(p.PrometheusConfig)
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

func addNodeMetricToNodesMetricMap(nodesMetricMap *DaoMetricTypes.NodeMetricMap, nodeMetricChan <-chan DaoMetricTypes.NodeMetric, done chan<- bool) {
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
