package node

import (
	"context"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"fmt"

	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

func SyncWithDatahub(client client.Client, conn *grpc.ClientConn) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	nodeList := corev1.NodeList{}
	if err := client.List(ctx, nil, &nodeList); err != nil {
		return errors.Errorf(
			"Sync nodes with datahub failed due to list nodes from cluster failed: %s", err.Error())
	}
	datahubNodeRepo := NewNodeRepository(conn)
	if len(nodeList.Items) > 0 {
		if err := datahubNodeRepo.CreateNodes(nodeList.Items); err != nil {
			return fmt.Errorf(
				"Sync nodes with datahub failed due to register node failed: %s", err.Error())
		}
	}

	// Clean up unexisting nodes from Datahub
	existingNodeMap := make(map[string]bool)
	for _, node := range nodeList.Items {
		existingNodeMap[node.Name] = true
	}

	nodesFromDatahub, err := datahubNodeRepo.ListNodes()
	if err != nil {
		return fmt.Errorf(
			"Sync nodes with datahub failed due to list nodes from datahub failed: %s", err.Error())
	}
	nodesNeedDeleting := make([]*datahub_resources.Node, 0)
	for _, n := range nodesFromDatahub {
		if _, exist := existingNodeMap[n.ObjectMeta.GetName()]; exist {
			continue
		}
		nodesNeedDeleting = append(nodesNeedDeleting, n)
	}
	if len(nodesNeedDeleting) > 0 {
		err = datahubNodeRepo.DeleteNodes(nodesNeedDeleting)
		if err != nil {
			return errors.Wrap(err, "delete nodes from Datahub failed")
		}
	}

	return nil
}
