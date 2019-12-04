package namespace

import (
	"context"
	"fmt"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	k8sutils "github.com/containers-ai/alameda/pkg/utils/kubernetes"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

func SyncWithDatahub(client client.Client, conn *grpc.ClientConn) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	namespaceList := corev1.NamespaceList{}
	if err := client.List(ctx, &namespaceList); err != nil {
		return errors.Errorf(
			"Sync namespaces with datahub failed due to list namespaces from cluster failed: %s", err.Error())
	}

	clusterUID, err := k8sutils.GetClusterUID(client)
	if err != nil {
		return errors.Wrap(err, "get cluster uid failed")
	}

	datahubNamespaceRepo := NewNamespaceRepository(conn, clusterUID)
	if len(namespaceList.Items) > 0 {
		if err := datahubNamespaceRepo.CreateNamespaces(namespaceList.Items); err != nil {
			return fmt.Errorf(
				"Sync namespaces with datahub failed due to register namespace failed: %s", err.Error())
		}
	}

	// Clean up unexisting namespaces from Datahub
	existingNamespaceMap := make(map[string]bool)
	for _, namespace := range namespaceList.Items {
		existingNamespaceMap[namespace.GetName()] = true
	}

	namespacesFromDatahub, err := datahubNamespaceRepo.ListNamespaces()
	if err != nil {
		return fmt.Errorf(
			"Sync namespaces with datahub failed due to list namespaces from datahub failed: %s", err.Error())
	}
	namespacesNeedDeleting := make([]*datahub_resources.Namespace, 0)
	for _, n := range namespacesFromDatahub {
		if datahubNamespaceRepo.IsNSExcluded(n.GetObjectMeta().GetName()) {
			namespacesNeedDeleting = append(namespacesNeedDeleting, n)
			continue
		}
		if _, exist := existingNamespaceMap[n.GetObjectMeta().GetName()]; exist {
			continue
		}
		namespacesNeedDeleting = append(namespacesNeedDeleting, n)
	}
	if len(namespacesNeedDeleting) > 0 {
		err = datahubNamespaceRepo.DeleteNamespaces(namespacesNeedDeleting)
		if err != nil {
			return errors.Wrap(err, "delete namespaces from Datahub failed")
		}
	}

	return nil
}
