package main

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	datahub_client "github.com/containers-ai/alameda/operator/datahub/client"
	datahub_node "github.com/containers-ai/alameda/operator/datahub/client/node"
	datahub_pod "github.com/containers-ai/alameda/operator/datahub/client/pod"
	"github.com/containers-ai/alameda/operator/pkg/utils/resources"
	alamutils "github.com/containers-ai/alameda/pkg/utils"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func registerNodes(client client.Client, retryInterval int64) {
	scope.Info("Start registering nodes to datahub.")
	for {
		time.Sleep(time.Duration(retryInterval) * time.Second)
		if err := startRegisteringNodes(client); err == nil {
			scope.Info("Register nodes to datahub successfully.")
			break
		} else {
			scope.Errorf("Register nodes to datahub failed due to %s.", err.Error())
		}
	}
	scope.Info("Registering nodes to datahub is done.")
}

func startRegisteringNodes(client client.Client) error {
	listResources := resources.NewListResources(client)
	nodes, err := listResources.ListAllNodes()

	if err != nil {
		return fmt.Errorf("register nodes to Datahub failed: %s", err.Error())
	}

	if len(nodes) == 0 {
		return fmt.Errorf("No nodes found to register to datahub")
	}

	scope.Infof(fmt.Sprintf("%v nodes found in cluster.", len(nodes)))
	datahubNodeRepo := datahub_node.NewAlamedaNodeRepository()
	err = datahubNodeRepo.CreateAlamedaNode(nodes)
	if err != nil {
		return err
	}
	nodesToDel := []*corev1.Node{}
	alamNodes, err := datahubNodeRepo.ListAlamedaNodes()
	for _, alamNode := range alamNodes {
		toDel := true
		for _, node := range nodes {
			if node.GetName() == alamNode.GetName() {
				toDel = false
				break
			}
		}
		if !toDel {
			continue
		}
		delNode := &corev1.Node{}
		delNode.SetName(alamNode.GetName())
		nodesToDel = append(nodesToDel, delNode)
	}

	if len(nodesToDel) > 0 {
		scope.Debugf("Nodes removed from datahub. %s", alamutils.InterfaceToString(nodesToDel))
		err := datahubNodeRepo.DeleteAlamedaNodes(nodesToDel)
		if err != nil {
			return err
		}
	}

	return err
}

func syncAlamedaPodsWithDatahub(client client.Client, retryInterval int64) {
	scope.Info("Start syncing alameda pods to datahub.")
	for {
		if err := startSyncingAlamedaPodsWithDatahubSuccess(client); err == nil {
			scope.Info("Sync alameda pod with datahub successfully.")
			break
		} else {
			scope.Errorf("Sync alameda pod with datahub failed due to %s", err.Error())
		}
		time.Sleep(time.Duration(retryInterval) * time.Second)
	}
	scope.Info("Syncing alameda pods to datahub is done.")
}

func startSyncingAlamedaPodsWithDatahubSuccess(client client.Client) error {
	datahubPodRepo := datahub_pod.NewPodRepository()
	alamedaPods, err := datahubPodRepo.ListAlamedaPods()
	if err != nil {
		return fmt.Errorf("Sync alameda pod with datahub failed: %s", err.Error())
	}
	podsNeedToRm := []*datahub_v1alpha1.Pod{}
	getResource := resources.NewGetResource(client)

	for _, alamedaPod := range alamedaPods {
		namespacedName := alamedaPod.GetNamespacedName()
		alamPodNS := namespacedName.GetNamespace()
		alamPodName := namespacedName.GetName()
		_, err := getResource.GetPod(alamPodNS, alamPodName)
		if err != nil && k8sErrors.IsNotFound(err) {
			podsNeedToRm = append(podsNeedToRm, alamedaPod)
			continue
		} else if err != nil {
			return fmt.Errorf("Get pod (%s/%s) failed while sync alameda pod with datahub. (%s)", alamPodNS, alamPodName, err.Error())
		}

		alamedaScaler := alamedaPod.GetAlamedaScaler()
		alamScalerNS := alamedaScaler.GetNamespace()
		alamScalerName := alamedaScaler.GetName()
		_, err = getResource.GetAlamedaScaler(alamScalerNS, alamScalerName)
		if err != nil && k8sErrors.IsNotFound(err) {
			podsNeedToRm = append(podsNeedToRm, alamedaPod)
			continue
		} else if err != nil {
			return fmt.Errorf("Get alameda scaler (%s/%s) failed while sync alameda pod with datahub. (%s)", alamedaScaler, alamScalerName, err.Error())
		}
	}

	if len(podsNeedToRm) > 0 {
		err := datahubPodRepo.DeletePods(podsNeedToRm)
		return err
	}
	return nil
}

func syncAlamedaResourcesWithDatahub(client client.Client, retryInterval int64) {
	for {
		err := startSyncingAlamedaResourcesWithDatahubSuccess(client)
		if err == nil {
			return
		}
		scope.Errorf("Sync AlamedaResource with datahub failed due to %s", err.Error())
		time.Sleep(time.Duration(retryInterval) * time.Second)
	}
}

// delete controllers that are not existed in the cluster and their AlamedaScaler owner is not existed
func startSyncingAlamedaResourcesWithDatahubSuccess(client client.Client) error {

	// Get current conrtoller list from Alameda-Datahub
	datahubK8SResourceRepository := datahub_client.NewK8SResource()
	alamedaResources, err := datahubK8SResourceRepository.ListAlamedaWatchedResource(nil)
	if err != nil {
		return errors.Wrap(err, "list resources watched by Alameda failed")
	}

	controllersNeedToRm := []*datahub_v1alpha1.Controller{}
	getResource := resources.NewGetResource(client)
	for _, alamedaResource := range alamedaResources {

		info := alamedaResource.GetControllerInfo()
		if info == nil {
			continue
		}
		resourceNamespace := ""
		resourceName := ""
		if namespacedName := info.GetNamespacedName(); namespacedName == nil {
			continue
		} else {
			resourceNamespace = namespacedName.GetNamespace()
			resourceName = namespacedName.GetName()
		}

		// Get controller from k8s, if controller is not existed, append conttoller to the list that needs to delete
		kind := info.GetKind()
		switch kind {
		case datahub_v1alpha1.Kind_DEPLOYMENT:
			_, err := getResource.GetDeployment(resourceNamespace, resourceName)
			if err != nil && k8sErrors.IsNotFound(err) {
				controllersNeedToRm = append(controllersNeedToRm, alamedaResource)
				continue
			} else if err != nil {
				return errors.Wrapf(err, "get Deployment (%s/%s) failed", resourceNamespace, resourceName)
			}
		case datahub_v1alpha1.Kind_DEPLOYMENTCONFIG:
			_, err := getResource.GetDeploymentConfig(resourceNamespace, resourceName)
			if err != nil && k8sErrors.IsNotFound(err) {
				controllersNeedToRm = append(controllersNeedToRm, alamedaResource)
				continue
			} else if err != nil {
				return errors.Wrapf(err, "get DeploymentConfig (%s/%s) failed", resourceNamespace, resourceName)
			}
		case datahub_v1alpha1.Kind_STATEFULSET:
			_, err := getResource.GetStatefulSet(resourceNamespace, resourceName)
			if err != nil && k8sErrors.IsNotFound(err) {
				controllersNeedToRm = append(controllersNeedToRm, alamedaResource)
				continue
			} else if err != nil {
				return errors.Wrapf(err, "get StatefulSet (%s/%s) failed", resourceNamespace, resourceName)
			}
		default:
			return errors.Errorf("unknown controller datahub kind \"%s\"", kind.String())
		}

		// Get AlamedaScaler that owning this controller from k8s,
		// if AlamedaScaler is not existed, append conttoller to the list that needs to delete
		ownerInfos := alamedaResource.GetOwnerInfo()
		for _, ownerInfo := range ownerInfos {
			if ownerInfo == nil {
				continue
			}
			if ownerInfo.GetKind() != datahub_v1alpha1.Kind_ALAMEDASCALER {
				continue
			}
			alamedaScalerNamespacedName := ownerInfo.GetNamespacedName()
			if alamedaScalerNamespacedName == nil {
				continue
			}
			alamScalerNS := alamedaScalerNamespacedName.GetNamespace()
			alamScalerName := alamedaScalerNamespacedName.GetName()
			_, err = getResource.GetAlamedaScaler(alamScalerNS, alamScalerName)
			if err != nil && k8sErrors.IsNotFound(err) {
				controllersNeedToRm = append(controllersNeedToRm, alamedaResource)
				break
			} else if err != nil {
				return errors.Wrapf(err, "get AlamedaScaler (%s/%s) failed", alamScalerNS, alamScalerName)
			}
		}
	}

	if len(controllersNeedToRm) > 0 {
		err := datahubK8SResourceRepository.DeleteAlamedaWatchedResource(controllersNeedToRm)
		return errors.Wrap(err, "delete resources watched by Alameda from Alameda-Datahub failed")
	}
	return nil
}
