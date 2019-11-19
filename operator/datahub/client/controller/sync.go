package controller

import (
	"context"
	"time"

	k8sutils "github.com/containers-ai/alameda/pkg/utils/kubernetes"
	appsapi_v1 "github.com/openshift/api/apps/v1"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	appsv1 "k8s.io/api/apps/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"fmt"

	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

func SyncWithDatahub(client client.Client, conn *grpc.ClientConn) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Clean up unexisting controllers from Datahub
	existingControllerMap := make(map[string]bool)

	deploymentList := appsv1.DeploymentList{}
	if err := client.List(ctx, &deploymentList); err != nil {
		return errors.Errorf(
			"Sync controller with datahub failed due to list deployments from cluster failed: %s",
			err.Error())
	}
	for _, controller := range deploymentList.Items {
		existingControllerMap[fmt.Sprintf("%s/%s/%s", datahub_resources.Kind_DEPLOYMENT.String(),
			controller.GetNamespace(), controller.GetName())] = true
	}

	statefulSetList := appsv1.StatefulSetList{}
	if err := client.List(ctx, &statefulSetList); err != nil {
		return errors.Errorf(
			"Sync controller with datahub failed due to list statefulsets from cluster failed: %s",
			err.Error())
	}
	for _, controller := range statefulSetList.Items {
		existingControllerMap[fmt.Sprintf("%s/%s/%s", datahub_resources.Kind_STATEFULSET.String(),
			controller.GetNamespace(), controller.GetName())] = true
	}

	deploymentConfigList := appsapi_v1.DeploymentConfigList{}
	if err := client.List(ctx, &deploymentConfigList); err != nil {
		return errors.Errorf(
			"Sync controller with datahub failed due to list deploymentconfigs from cluster failed: %s",
			err.Error())
	}
	for _, controller := range deploymentConfigList.Items {
		existingControllerMap[fmt.Sprintf("%s/%s/%s", datahub_resources.Kind_DEPLOYMENTCONFIG.String(),
			controller.GetNamespace(), controller.GetName())] = true
	}

	clusterUID, err := k8sutils.GetClusterUID(client)
	if err != nil {
		return errors.Wrap(err, "get cluster uid failed")
	}

	datahubControllerRepo := NewControllerRepository(conn, clusterUID)
	controllersFromDatahub, err := datahubControllerRepo.ListControllers()
	if err != nil {
		return fmt.Errorf(
			"Sync controllers with datahub failed due to list controllers from datahub failed: %s",
			err.Error())
	}

	deploymentsNeedDeleting := make([]*datahub_resources.Controller, 0)
	statefulSetsNeedDeleting := make([]*datahub_resources.Controller, 0)
	deploymentConfigsNeedDeleting := make([]*datahub_resources.Controller, 0)
	for _, n := range controllersFromDatahub {
		if _, exist := existingControllerMap[fmt.Sprintf("%s/%s/%s", n.GetKind().String(),
			n.GetObjectMeta().GetNamespace(), n.GetObjectMeta().GetName())]; exist {
			continue
		}

		if n.GetKind() == datahub_resources.Kind_DEPLOYMENT {
			deploymentsNeedDeleting = append(deploymentsNeedDeleting, n)
		}
		if n.GetKind() == datahub_resources.Kind_STATEFULSET {
			statefulSetsNeedDeleting = append(statefulSetsNeedDeleting, n)
		}
		if n.GetKind() == datahub_resources.Kind_DEPLOYMENTCONFIG {
			deploymentConfigsNeedDeleting = append(deploymentConfigsNeedDeleting, n)
		}
	}

	if len(deploymentsNeedDeleting) > 0 {
		err = datahubControllerRepo.DeleteControllers(ctx, deploymentsNeedDeleting, nil)
		if err != nil {
			return errors.Wrap(err, "delete deployments from Datahub failed")
		}
	}
	if len(statefulSetsNeedDeleting) > 0 {
		err = datahubControllerRepo.DeleteControllers(ctx, statefulSetsNeedDeleting, nil)
		if err != nil {
			return errors.Wrap(err, "delete statefulset from Datahub failed")
		}
	}
	if len(deploymentConfigsNeedDeleting) > 0 {
		err = datahubControllerRepo.DeleteControllers(ctx, deploymentConfigsNeedDeleting, nil)
		if err != nil {
			return errors.Wrap(err, "delete deploymentconfig from Datahub failed")
		}
	}
	return nil
}
