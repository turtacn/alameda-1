package metric

import (
	"github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/nodeMemoryBytesTotal"
	"github.com/containers-ai/alameda/datahub/pkg/entity/prometheus/nodeMemoryUtilization"
	"github.com/containers-ai/alameda/datahub/pkg/repository/prometheus"
	"github.com/containers-ai/alameda/pkg/utils/log"
	"github.com/pkg/errors"
)

var (
	scope = log.RegisterScope("node memory usage bytes", "node memory usage bytes", 0)
)

// NodeMemoryUsageBytesRepository Repository to access metric from prometheus
type NodeMemoryUsageBytesRepository struct {
	PrometheusConfig prometheus.Config
}

// NewNodeMemoryUsageBytesRepositoryWithConfig New node cpu usage percentage repository with prometheus configuration
func NewNodeMemoryUsageBytesRepositoryWithConfig(cfg prometheus.Config) NodeMemoryUsageBytesRepository {
	return NodeMemoryUsageBytesRepository{PrometheusConfig: cfg}
}

// ListMetricsByNodeName Provide metrics from response of querying request contain namespace, pod_name and default labels
func (n NodeMemoryUsageBytesRepository) ListMetricsByNodeName(nodeName string, options ...Option) ([]prometheus.Entity, error) {

	var (
		nodeMemoryUsageBytesEntities = make([]prometheus.Entity, 0)
	)

	nodeMemoryBytesTotalRepository := NewNodeMemoryBytesTotalRepositoryWithConfig(n.PrometheusConfig)
	nodeMemoryUtilizationRepository := NewNodeMemoryUtilizationRepositoryWithConfig(n.PrometheusConfig)

	errChan := make(chan error)
	entitiesChan := make(chan []prometheus.Entity)
	fetchFunctions := []nodeMetricsFetchingFunction{nodeMemoryBytesTotalRepository.ListMetricsByNodeName, nodeMemoryUtilizationRepository.ListMetricsByNodeName}
	entitiesSlice := make([][]prometheus.Entity, len(fetchFunctions))
	for _, f := range fetchFunctions {
		go n.fetchNodeMetricsFromFunction(f, entitiesChan, errChan, nodeName, options...)
	}

	goroutineDoneIndex := 0
	for goroutineDoneIndex < len(fetchFunctions) {
		select {
		case entities := <-entitiesChan:
			entitiesSlice[goroutineDoneIndex] = entities
		case err := <-errChan:
			return nodeMemoryUsageBytesEntities, errors.Wrap(err, "list node memory usage bytes by node name failed")
		}
		goroutineDoneIndex++
	}

	nodeMemoryUsageBytesEntities, err := n.buildEntitiesFromNodeMemoryBytesTotalEntitesMultiplyNodeMemoryUtilizationEntities(entitiesSlice[0], entitiesSlice[1])
	if err != nil {
		return nodeMemoryUsageBytesEntities, errors.Wrap(err, "list node memory usage bytes by node name failed")
	}

	return nodeMemoryUsageBytesEntities, nil
}

func (n NodeMemoryUsageBytesRepository) fetchNodeMetricsFromFunction(fetchFunction nodeMetricsFetchingFunction, entitiesChan chan []prometheus.Entity, errorChan chan error, nodeName string, options ...Option) {
	entities, err := fetchFunction(nodeName, options...)
	if err != nil {
		errorChan <- err
	} else {
		entitiesChan <- entities
	}
}

func (n NodeMemoryUsageBytesRepository) buildEntitiesFromNodeMemoryBytesTotalEntitesMultiplyNodeMemoryUtilizationEntities(firstEntites, secondEntities []prometheus.Entity) ([]prometheus.Entity, error) {

	var (
		nodeMemoryUsageBytesEntities = make([]prometheus.Entity, 0)
		nodeMemoryBytesTotaEntityMap = make(map[string]prometheus.Entity)
	)

	if len(firstEntites) != len(secondEntities) {
		return nodeMemoryUsageBytesEntities, errors.Errorf("build entities failed: length of nodeMemoryBytesTotal entities is not equal to length of nodeMemoryUtilization entities")
	}

	for _, entity := range firstEntites {
		if nodeName, exist := entity.Labels[nodeMemoryBytesTotal.NodeLabel]; !exist {
			return nodeMemoryUsageBytesEntities, errors.Errorf("build entities failed: node label %s doe not exist in entity's labels,received entity: %+v", nodeMemoryBytesTotal.NodeLabel, entity)
		} else {
			nodeMemoryBytesTotaEntityMap[nodeName] = entity
		}
	}

	for _, entity := range secondEntities {
		nodeName, exist := entity.Labels[nodeMemoryUtilization.NodeLabel]
		if !exist {
			return nodeMemoryUsageBytesEntities, errors.Errorf("build entities failed: node label %s doe not exist in entity's labels,received entity: %+v", nodeMemoryBytesTotal.NodeLabel, entity)
		}
		firstEntity, exist := nodeMemoryBytesTotaEntityMap[nodeName]
		if !exist {
			return nodeMemoryUsageBytesEntities, errors.Errorf("build entities failed: node name %s does not present in both nodeMemoryUtilization and nodeMemoryBytesTotal entities", nodeName)
		}

		options := []operateOption{
			operateOptionsMatchType(matchTypeOn),
			operateOptionsLabels([]string{nodeMemoryUtilization.NodeLabel}),
		}
		usageBytesEntity, err := oneToOneMultiply(firstEntity, entity, options...)
		if err != nil {
			return nodeMemoryUsageBytesEntities, errors.Wrap(err, "build entities failed:")
		}
		nodeMemoryUsageBytesEntities = append(nodeMemoryUsageBytesEntities, usageBytesEntity)
	}

	return nodeMemoryUsageBytesEntities, nil

}
