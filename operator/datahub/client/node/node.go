package node

import (
	"context"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	"github.com/containers-ai/alameda/operator/datahub/client"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

// providerID: aws:///us-west-2a/i-0769ec8570198bf4b --> <provider_raw>//<region>//<instance_id>

// AlamedaNodeRepository creates predicted node to datahub
type AlamedaNodeRepository struct {
	conn          *grpc.ClientConn
	datahubClient datahub_v1alpha1.DatahubServiceClient

	clusterUID string
}

// NewNodeRepository return AlamedaNodeRepository instance
func NewNodeRepository(conn *grpc.ClientConn, clusterUID string) *AlamedaNodeRepository {

	datahubClient := datahub_v1alpha1.NewDatahubServiceClient(conn)

	return &AlamedaNodeRepository{
		conn:          conn,
		datahubClient: datahubClient,

		clusterUID: clusterUID,
	}
}

func (repo *AlamedaNodeRepository) Close() {
	repo.conn.Close()
}

// CreateNodes creates predicted node to datahub
func (repo *AlamedaNodeRepository) CreateNodes(nodes []*datahub_resources.Node) error {

	if len(nodes) > 0 {
		req := datahub_resources.CreateNodesRequest{
			Nodes: nodes,
		}
		if resp, err := repo.datahubClient.CreateNodes(context.Background(), &req); err != nil {
			return errors.Wrap(err, "create nodes to datahub failed")
		} else if _, err := client.IsResponseStatusOK(resp); err != nil {
			return errors.Wrap(err, "create nodes to Datahub failed")
		}
	}
	return nil
}

// DeleteNodes delete predicted node from datahub
func (repo *AlamedaNodeRepository) DeleteNodes(arg interface{}) error {
	objMeta := []*datahub_resources.ObjectMeta{}
	if nodes, ok := arg.([]*datahub_resources.Node); ok {
		for _, node := range nodes {
			copyNode := *node
			objMeta = append(objMeta, copyNode.ObjectMeta)
		}
	}
	if meta, ok := arg.([]*datahub_resources.ObjectMeta); ok {
		objMeta = meta
	}

	req := datahub_resources.DeleteNodesRequest{
		ObjectMeta: objMeta,
	}

	if resp, err := repo.datahubClient.DeleteNodes(context.Background(), &req); err != nil {
		return errors.Wrap(err, "delete node from Datahub failed")
	} else if _, err := client.IsResponseStatusOK(resp); err != nil {
		return errors.Wrap(err, "delete nodes from Datahub failed")
	}
	return nil
}

// ListNodes lists nodes to datahub
func (repo *AlamedaNodeRepository) ListNodes() ([]*datahub_resources.Node, error) {
	return repo.listAlamedaNodes()
}

func (repo *AlamedaNodeRepository) listAlamedaNodes() ([]*datahub_resources.Node, error) {
	req := datahub_resources.ListNodesRequest{
		ObjectMeta: []*datahub_resources.ObjectMeta{
			&datahub_resources.ObjectMeta{
				ClusterName: repo.clusterUID,
			},
		},
	}
	resp, err := repo.datahubClient.ListNodes(context.Background(), &req)
	if err != nil {
		return nil, errors.Errorf("list nodes from Datahub failed: %s", err.Error())
	} else if resp == nil {
		return nil, errors.Errorf("list nodes from Datahub failed, receive nil response")
	} else if _, err := client.IsResponseStatusOK(resp.Status); err != nil {
		return nil, errors.Wrap(err, "list nodes from Datahub failed")
	}
	return resp.Nodes, nil
}
