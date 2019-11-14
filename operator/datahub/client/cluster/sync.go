package cluster

import (
	"context"
	"fmt"
	"time"

	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func SyncWithDatahub(client client.Client, conn *grpc.ClientConn) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cm := corev1.ConfigMap{}
	if err := client.Get(ctx, types.NamespacedName{
		Namespace: "default",
		Name:      "cluster-info",
	}, &cm); err != nil {
		return errors.Errorf(
			"Sync cluster with datahub failed due to get cluster from cluster failed: %s",
			err.Error())
	}
	datahubClusterRepo := NewClusterRepository(conn)

	if err := datahubClusterRepo.CreateClusters([]*datahub_resources.Cluster{
		&datahub_resources.Cluster{
			ObjectMeta: &datahub_resources.ObjectMeta{
				Name: string(cm.GetUID()),
			},
		},
	}); err != nil {
		return fmt.Errorf(
			"Sync cluster with datahub failed due to register cluster failed: %s",
			err.Error())
	}

	return nil
}
