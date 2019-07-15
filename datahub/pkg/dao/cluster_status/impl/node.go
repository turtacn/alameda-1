package impl

import (
	DaoClusterStatus "github.com/containers-ai/alameda/datahub/pkg/dao/cluster_status"
	RepoInfluxClusterStatus "github.com/containers-ai/alameda/datahub/pkg/repository/influxdb/cluster_status"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"github.com/pkg/errors"
)

// Implement Node interface
type Node struct {
	InfluxDBConfig InternalInflux.Config
}

func (node *Node) RegisterAlamedaNodes(alamedaNodes []*datahub_v1alpha1.Node) error {
	nodeRepository := RepoInfluxClusterStatus.NewNodeRepository(&node.InfluxDBConfig)
	return nodeRepository.AddAlamedaNodes(alamedaNodes)
}

func (node *Node) DeregisterAlamedaNodes(alamedaNodes []*datahub_v1alpha1.Node) error {
	nodeRepository := RepoInfluxClusterStatus.NewNodeRepository(&node.InfluxDBConfig)
	return nodeRepository.RemoveAlamedaNodes(alamedaNodes)
}

func (node *Node) ListAlamedaNodes(timeRange *datahub_v1alpha1.TimeRange) ([]*datahub_v1alpha1.Node, error) {
	alamedaNodes := []*datahub_v1alpha1.Node{}
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

func (node *Node) ListNodes(request DaoClusterStatus.ListNodesRequest) ([]*datahub_v1alpha1.Node, error) {
	nodes := []*datahub_v1alpha1.Node{}
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
