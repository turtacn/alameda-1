package impl

import (
	cluster_status_dao "github.com/containers-ai/alameda/datahub/pkg/dao/cluster_status"
	influxdb_entity_cluster_status "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/cluster_status"
	influxdb_repository "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	influxdb_repository_cluster_status "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb/cluster_status"
	"github.com/containers-ai/alameda/pkg/utils/log"
)

var (
	scope = log.RegisterScope("node_dao_implement", "node dao implement", 0)
)

// Implement Node interface
type Node struct {
	InfluxDBConfig influxdb_repository.Config
}

func (node *Node) RegisterAlamedaNodes(alamedaNodes []*cluster_status_dao.AlamedaNode) error {
	nodeRepository := influxdb_repository_cluster_status.NewNodeRepository(&node.InfluxDBConfig)
	return nodeRepository.AddAlamedaNodes(alamedaNodes)
}

func (node *Node) DeregisterAlamedaNodes(alamedaNodes []*cluster_status_dao.AlamedaNode) error {
	nodeRepository := influxdb_repository_cluster_status.NewNodeRepository(&node.InfluxDBConfig)
	return nodeRepository.RemoveAlamedaNodes(alamedaNodes)
}

func (node *Node) ListAlamedaNodes() ([]*cluster_status_dao.AlamedaNode, error) {
	alamedaNodes := []*cluster_status_dao.AlamedaNode{}
	nodeRepository := influxdb_repository_cluster_status.NewNodeRepository(&node.InfluxDBConfig)
	entities, _ := nodeRepository.ListAlamedaNodes()
	for _, entity := range entities {
		alamedaNodes = append(alamedaNodes, &cluster_status_dao.AlamedaNode{
			Name: entity.Fields[string(influxdb_entity_cluster_status.NodeName)].(string),
		})
	}
	return alamedaNodes, nil
}
