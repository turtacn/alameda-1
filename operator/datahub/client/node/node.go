package node

import (
	"context"

	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
)

// providerID: aws:///us-west-2a/i-0769ec8570198bf4b --> <provider_raw>//<region>//<instance_id>

// AlamedaNodeRepository creates predicted node to datahub
type AlamedaNodeRepository struct {
	conn          *grpc.ClientConn
	datahubClient datahub_v1alpha1.DatahubServiceClient
}

// NewAlamedaNodeRepository return AlamedaNodeRepository instance
func NewAlamedaNodeRepository(conn *grpc.ClientConn) *AlamedaNodeRepository {

	datahubClient := datahub_v1alpha1.NewDatahubServiceClient(conn)

	return &AlamedaNodeRepository{
		conn:          conn,
		datahubClient: datahubClient,
	}
}

func (repo *AlamedaNodeRepository) Close() {
	repo.conn.Close()
}

// CreateAlamedaNode creates predicted node to datahub
func (repo *AlamedaNodeRepository) CreateAlamedaNode(nodes []*datahub_resources.Node) error {
	return repo.createAlamedaNode(nodes)
}

func (repo *AlamedaNodeRepository) createAlamedaNode(nodes []*datahub_resources.Node) error {

	req := datahub_resources.CreateNodesRequest{
		Nodes: nodes,
	}

	if reqRes, err := repo.datahubClient.CreateNodes(context.Background(), &req); err != nil {
		return errors.Errorf("create nodes to datahub failed: %s", err.Error())
	} else if reqRes == nil {
		return errors.Errorf("create nodes to datahub failed: receive nil status")
	} else if reqRes.Code != int32(code.Code_OK) {
		return errors.Errorf("create nodes to datahub failed: receive statusCode: %d, message: %s", reqRes.Code, reqRes.Message)
	}
	return nil
}

// DeleteAlamedaNodes delete predicted node from datahub
func (repo *AlamedaNodeRepository) DeleteAlamedaNodes(arg interface{}) error {
	objMeta := []*datahub_resources.ObjectMeta{}
	if nodes, ok := arg.([]*datahub_resources.Node); ok {
		for _, node := range nodes {
			objMeta = append(objMeta, &datahub_resources.ObjectMeta{
				Name: node.ObjectMeta.GetName(),
			})
		}
	}
	if meta, ok := arg.([]*datahub_resources.ObjectMeta); ok {
		objMeta = meta
	}

	req := datahub_resources.DeleteNodesRequest{
		ObjectMeta: objMeta,
	}

	if resp, err := repo.datahubClient.DeleteNodes(context.Background(), &req); err != nil {
		return errors.Errorf("delete node from Datahub failed: %s", err.Error())
	} else if resp.Code != int32(code.Code_OK) {
		return errors.Errorf("delete node from Datahub failed: receive code: %d, message: %s", resp.Code, resp.Message)
	}
	return nil
}

// ListAlamedaNodes lists nodes to datahub
func (repo *AlamedaNodeRepository) ListAlamedaNodes() ([]*datahub_resources.Node, error) {
	return repo.listAlamedaNodes()
}

func (repo *AlamedaNodeRepository) listAlamedaNodes() ([]*datahub_resources.Node, error) {
	alamNodes := []*datahub_resources.Node{}
	req := datahub_resources.ListNodesRequest{}
	if reqRes, err := repo.datahubClient.ListNodes(context.Background(), &req); err != nil {
		if reqRes.Status != nil {
			return alamNodes, errors.Errorf("list nodes from Datahub failed: %s", err.Error())
		}
		return alamNodes, err
	} else {
		alamNodes = reqRes.GetNodes()
	}
	return alamNodes, nil
}
