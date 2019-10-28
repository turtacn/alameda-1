package influxdb

import (
	DaoClusterStatus "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus"
	RepoInfluxClusterStatus "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/clusterstatus"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/pkg/errors"
)

// Implement Node interface
type Node struct {
	InfluxDBConfig InternalInflux.Config
}

func (node *Node) RegisterAlamedaNodes(alamedaNodes []*ApiResources.Node) error {
	nodeRepository := RepoInfluxClusterStatus.NewNodeRepository(&node.InfluxDBConfig)
	return nodeRepository.AddAlamedaNodes(alamedaNodes)
}

func (node *Node) DeregisterAlamedaNodes(alamedaNodes []*ApiResources.Node) error {
	nodeRepository := RepoInfluxClusterStatus.NewNodeRepository(&node.InfluxDBConfig)
	return nodeRepository.RemoveAlamedaNodes(alamedaNodes)
}

func (node *Node) ListAlamedaNodes(timeRange *ApiCommon.TimeRange) ([]*ApiResources.Node, error) {
	alamedaNodes := make([]*ApiResources.Node, 0)
	nodeRepository := RepoInfluxClusterStatus.NewNodeRepository(&node.InfluxDBConfig)
	entities, err := nodeRepository.ListAlamedaNodes(timeRange)
	if err != nil {
		return alamedaNodes, errors.Wrap(err, "list alameda nodes failed")
	}
	for _, entity := range entities {
		alamedaNodes = append(alamedaNodes, entity.BuildDatahubNode())
	}
	return alamedaNodes, nil
}

func (node *Node) ListNodes(request DaoClusterStatus.ListNodesRequest) ([]*ApiResources.Node, error) {
	nodes := make([]*ApiResources.Node, 0)
	nodeRepository := RepoInfluxClusterStatus.NewNodeRepository(&node.InfluxDBConfig)
	entities, err := nodeRepository.ListNodes(request)
	if err != nil {
		return nodes, errors.Wrap(err, "list nodes failed")
	}
	for _, entity := range entities {
		nodes = append(nodes, entity.BuildDatahubNode())
	}
	return nodes, nil
}
