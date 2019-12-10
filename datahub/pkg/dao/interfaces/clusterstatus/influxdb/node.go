package influxdb

import (
	DaoClusterTypes "github.com/containers-ai/alameda/datahub/pkg/dao/interfaces/clusterstatus/types"
	RepoInfluxCluster "github.com/containers-ai/alameda/datahub/pkg/dao/repositories/influxdb/clusterstatus"
	Metadata "github.com/containers-ai/alameda/datahub/pkg/kubernetes/metadata"
	InternalInflux "github.com/containers-ai/alameda/internal/pkg/database/influxdb"
)

// Implement Node interface
type Node struct {
	InfluxDBConfig InternalInflux.Config
}

func NewNodeWithConfig(config InternalInflux.Config) DaoClusterTypes.NodeDAO {
	return &Node{InfluxDBConfig: config}
}

func (p *Node) CreateNodes(nodes []*DaoClusterTypes.Node) error {
	nodeRepo := RepoInfluxCluster.NewNodeRepository(p.InfluxDBConfig)
	if err := nodeRepo.CreateNodes(nodes); err != nil {
		scope.Error(err.Error())
		return err
	}
	return nil
}

func (p *Node) ListNodes(request *DaoClusterTypes.ListNodesRequest) ([]*DaoClusterTypes.Node, error) {
	nodeRepo := RepoInfluxCluster.NewNodeRepository(p.InfluxDBConfig)
	nodes, err := nodeRepo.ListNodes(request)
	if err != nil {
		scope.Error(err.Error())
		return make([]*DaoClusterTypes.Node, 0), err
	}
	return nodes, nil
}

func (p *Node) DeleteNodes(request *DaoClusterTypes.DeleteNodesRequest) error {
	delPodsReq := p.genDeletePodsRequest(request)

	// Delete nodes
	nodeRepo := RepoInfluxCluster.NewNodeRepository(p.InfluxDBConfig)
	if err := nodeRepo.DeleteNodes(request); err != nil {
		scope.Error(err.Error())
		return err
	}

	// Delete pods
	podDAO := NewPodWithConfig(p.InfluxDBConfig)
	if err := podDAO.DeletePods(delPodsReq); err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}

func (p *Node) genDeletePodsRequest(request *DaoClusterTypes.DeleteNodesRequest) *DaoClusterTypes.DeletePodsRequest {
	delPodsReq := DaoClusterTypes.NewDeletePodsRequest()

	for _, objectMeta := range request.ObjectMeta {
		metadata := &Metadata.ObjectMeta{}
		metadata.NodeName = objectMeta.Name
		metadata.ClusterName = objectMeta.ClusterName

		podObjectMeta := DaoClusterTypes.NewPodObjectMeta(metadata, nil, nil, "", "")
		delPodsReq.PodObjectMeta = append(delPodsReq.PodObjectMeta, podObjectMeta)
	}

	return delPodsReq
}
