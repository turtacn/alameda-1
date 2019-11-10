package influxdb

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/clusterstatus"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

// Implement Node interface
type Node struct {
	InfluxDBConfig InternalInflux.Config
}

func NewNodeWithConfig(config InternalInflux.Config) DaoClusterTypes.NodeDAO {
	return &Node{InfluxDBConfig: config}
}

func (n *Node) CreateNodes(nodes []*DaoClusterTypes.Node) error {
	nodeRepo := RepoInfluxCluster.NewNodeRepository(&n.InfluxDBConfig)
	if err := nodeRepo.CreateNodes(nodes); err != nil {
		scope.Error(err.Error())
		return err
	}
	return nil
}

func (n *Node) ListNodes(request DaoClusterTypes.ListNodesRequest) ([]*DaoClusterTypes.Node, error) {
	nodeRepo := RepoInfluxCluster.NewNodeRepository(&n.InfluxDBConfig)
	nodes, err := nodeRepo.ListNodes(request)
	if err != nil {
		scope.Error(err.Error())
		return make([]*DaoClusterTypes.Node, 0), err
	}
	return nodes, nil
}

func (n *Node) DeleteNodes(nodes []*ApiResources.Node) error {
	nodeRepository := RepoInfluxCluster.NewNodeRepository(&n.InfluxDBConfig)
	return nodeRepository.DeleteNodes(nodes)
}

/*func (node *Node) RegisterAlamedaNodes(alamedaNodes []*ApiResources.Node) error {
	nodeRepository := RepoInfluxCluster.NewNodeRepository(&node.InfluxDBConfig)
	return nodeRepository.AddAlamedaNodes(alamedaNodes)
}

func (node *Node) ListAlamedaNodes(timeRange *ApiCommon.TimeRange) ([]*ApiResources.Node, error) {
	alamedaNodes := make([]*ApiResources.Node, 0)
	nodeRepository := RepoInfluxCluster.NewNodeRepository(&node.InfluxDBConfig)
	entities, err := nodeRepository.ListAlamedaNodes(timeRange)
	if err != nil {
		return alamedaNodes, errors.Wrap(err, "list alameda nodes failed")
	}
	for _, entity := range entities {
		alamedaNodes = append(alamedaNodes, entity.BuildDatahubNode())
	}
	return alamedaNodes, nil
}*/
