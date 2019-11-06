package influxdb

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/clusterstatus"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/pkg/errors"
)

// Implement Node interface
type Node struct {
	InfluxDBConfig InternalInflux.Config
}

func NewNodeWithConfig(config InternalInflux.Config) DaoClusterTypes.NodeDAO {
	return &Node{InfluxDBConfig: config}
}

func (node *Node) RegisterAlamedaNodes(alamedaNodes []*ApiResources.Node) error {
	nodeRepository := RepoInfluxCluster.NewNodeRepository(&node.InfluxDBConfig)
	return nodeRepository.AddAlamedaNodes(alamedaNodes)
}

func (node *Node) DeregisterAlamedaNodes(alamedaNodes []*ApiResources.Node) error {
	nodeRepository := RepoInfluxCluster.NewNodeRepository(&node.InfluxDBConfig)
	return nodeRepository.RemoveAlamedaNodes(alamedaNodes)
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
}

func (node *Node) ListNodes(request DaoClusterTypes.ListNodesRequest) ([]*ApiResources.Node, error) {
	nodes := make([]*ApiResources.Node, 0)
	nodeRepository := RepoInfluxCluster.NewNodeRepository(&node.InfluxDBConfig)
	entities, err := nodeRepository.ListNodes(request)
	if err != nil {
		return nodes, errors.Wrap(err, "list nodes failed")
	}
	for _, entity := range entities {
		nodes = append(nodes, entity.BuildDatahubNode())
	}
	return nodes, nil
}
