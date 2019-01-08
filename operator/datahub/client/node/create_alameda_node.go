package node

import (
	"context"

	datahubutils "github.com/containers-ai/alameda/operator/pkg/utils/datahub"
	logUtil "github.com/containers-ai/alameda/operator/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
)

var (
	createAlamedaNodeScope = logUtil.RegisterScope("create_alameda_node", "Create Alameda node.", 0)
)

// CreateAlamedaNode creates predicted node to datahub
type CreateAlamedaNode struct{}

// NewCreateAlamedaNode return CreateAlamedaNode instance
func NewCreateAlamedaNode() *CreateAlamedaNode {
	return &CreateAlamedaNode{}
}

// CreateAlamedaNode creates predicted node to datahub
func (createAlamedaNode *CreateAlamedaNode) CreateAlamedaNode(nodeList []corev1.Node) error {
	alamedaNodes := []*datahub_v1alpha1.Node{}
	for _, node := range nodeList {
		alamedaNodes = append(alamedaNodes, &datahub_v1alpha1.Node{
			Name: node.GetName(),
		})
	}
	req := datahub_v1alpha1.CreateAlamedaNodesRequest{
		AlamedaNodes: alamedaNodes,
	}
	conn, err := grpc.Dial(datahubutils.GetDatahubAddress(), grpc.WithInsecure())

	if err != nil {
		createAlamedaNodeScope.Error(err.Error())
		return err
	}

	defer conn.Close()
	aiServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(conn)
	if reqRes, err := aiServiceClnt.CreateAlamedaNodes(context.Background(), &req); err != nil {
		createAlamedaNodeScope.Error(reqRes.GetMessage())
		return err
	}
	return nil
}
