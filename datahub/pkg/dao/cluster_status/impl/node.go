package impl

import (
	influxdb_entity_cluster_status "github.com/containers-ai/alameda/datahub/pkg/entity/influxdb/cluster_status"
	influxdb_repository "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb"
	influxdb_repository_cluster_status "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb/cluster_status"
	"github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

var (
	scope = log.RegisterScope("node_dao_implement", "node dao implement", 0)
)

// Implement Node interface
type Node struct {
	InfluxDBConfig influxdb_repository.Config
}

func (node *Node) RegisterAlamedaNodes(alamedaNodes []*datahub_v1alpha1.Node) error {
	nodeRepository := influxdb_repository_cluster_status.NewNodeRepository(&node.InfluxDBConfig)
	return nodeRepository.AddAlamedaNodes(alamedaNodes)
}

func (node *Node) DeregisterAlamedaNodes(alamedaNodes []*datahub_v1alpha1.Node) error {
	nodeRepository := influxdb_repository_cluster_status.NewNodeRepository(&node.InfluxDBConfig)
	return nodeRepository.RemoveAlamedaNodes(alamedaNodes)
}

func (node *Node) ListAlamedaNodes() ([]*datahub_v1alpha1.Node, error) {
	alamedaNodes := []*datahub_v1alpha1.Node{}
	nodeRepository := influxdb_repository_cluster_status.NewNodeRepository(&node.InfluxDBConfig)
	entities, _ := nodeRepository.ListAlamedaNodes()
	for _, entity := range entities {
		alamedaNodes = append(alamedaNodes, &datahub_v1alpha1.Node{
			Name: entity.Fields[string(influxdb_entity_cluster_status.NodeName)].(string),
		})
	}
	return alamedaNodes, nil
}
