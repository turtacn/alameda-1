package main

import (
	"fmt"
	"time"

	"github.com/pkg/errors"

	datahub_client "github.com/containers-ai/alameda/operator/datahub/client"
	datahub_pod "github.com/containers-ai/alameda/operator/datahub/client/pod"
	"github.com/containers-ai/alameda/operator/pkg/utils/resources"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
	datahubPodRepo := datahub_pod.NewPodRepository(clusterUID)
	alamedaPods, err := datahubPodRepo.ListAlamedaPods()
	if err != nil {
		return fmt.Errorf("Sync alameda pod with datahub failed: %s", err.Error())
	}
	podsNeedToRm := []*datahub_resources.Pod{}
	getResource := resources.NewGetResource(client)

	for _, alamedaPod := range alamedaPods {
		alamPodNS := alamedaPod.GetObjectMeta().GetNamespace()
		alamPodName := alamedaPod.GetObjectMeta().GetName()
		_, err := getResource.GetPod(alamPodNS, alamPodName)
		if err != nil && k8sErrors.IsNotFound(err) {
			podsNeedToRm = append(podsNeedToRm, alamedaPod)
			continue
		} else if err != nil {
			return fmt.Errorf("Get pod (%s/%s) failed while sync alameda pod with datahub. (%s)", alamPodNS, alamPodName, err.Error())
		}

		alamedaScaler := alamedaPod.GetAlamedaPodSpec().GetAlamedaScaler()
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
	datahubK8SResourceRepository := datahub_client.NewK8SResource(clusterUID)
	alamedaResources, err := datahubK8SResourceRepository.ListAlamedaWatchedResource("", "")
	if err != nil {
		return errors.Wrap(err, "list resources watched by Alameda failed")
	}

	controllersNeedToRm := []*datahub_resources.Controller{}
	getResource := resources.NewGetResource(client)
	for _, alamedaResource := range alamedaResources {
		resourceNamespace := alamedaResource.GetObjectMeta().GetNamespace()
		resourceName := alamedaResource.GetObjectMeta().GetName()
		if resourceNamespace == "" && resourceName == "" {
			continue
		}

		// Get controller from k8s, if controller is not existed, append conttoller to the list that needs to delete
		kind := alamedaResource.GetKind()
		switch kind {
		case datahub_resources.Kind_DEPLOYMENT:
			_, err := getResource.GetDeployment(resourceNamespace, resourceName)
			if err != nil && k8sErrors.IsNotFound(err) {
				controllersNeedToRm = append(controllersNeedToRm, alamedaResource)
				continue
			} else if err != nil {
				return errors.Wrapf(err, "get Deployment (%s/%s) failed", resourceNamespace, resourceName)
			}
		case datahub_resources.Kind_DEPLOYMENTCONFIG:
			_, err := getResource.GetDeploymentConfig(resourceNamespace, resourceName)
			if err != nil && k8sErrors.IsNotFound(err) {
				controllersNeedToRm = append(controllersNeedToRm, alamedaResource)
				continue
			} else if err != nil {
				return errors.Wrapf(err, "get DeploymentConfig (%s/%s) failed", resourceNamespace, resourceName)
			}
		case datahub_resources.Kind_STATEFULSET:
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

		alamScalerNS := alamedaResource.GetAlamedaControllerSpec().GetAlamedaScaler().GetNamespace()
		alamScalerName := alamedaResource.GetAlamedaControllerSpec().GetAlamedaScaler().GetName()
		if alamScalerNS == "" && alamScalerName == "" {
			continue
		}
		_, err = getResource.GetAlamedaScaler(alamScalerNS, alamScalerName)
		if err != nil && k8sErrors.IsNotFound(err) {
			controllersNeedToRm = append(controllersNeedToRm, alamedaResource)
			break
		} else if err != nil {
			return errors.Wrapf(err, "get AlamedaScaler (%s/%s) failed", alamScalerNS, alamScalerName)
		}
	}

	if len(controllersNeedToRm) > 0 {
		err := datahubK8SResourceRepository.DeleteAlamedaWatchedResource(controllersNeedToRm)
		return errors.Wrap(err, "delete resources watched by Alameda from Alameda-Datahub failed")
	}
	return nil
}
