/*
Copyright 2019 The Alameda Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"

	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"

	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/api/v1alpha1"
	datahub_application "github.com/containers-ai/alameda/operator/datahub/client/application"
	datahub_controller "github.com/containers-ai/alameda/operator/datahub/client/controller"
	datahub_namespace "github.com/containers-ai/alameda/operator/datahub/client/namespace"
	datahub_pod "github.com/containers-ai/alameda/operator/datahub/client/pod"
	alamedascaler_reconciler "github.com/containers-ai/alameda/operator/pkg/reconciler/alamedascaler"
	"github.com/containers-ai/alameda/operator/pkg/utils"
	utilsresource "github.com/containers-ai/alameda/operator/pkg/utils/resources"
	alamutils "github.com/containers-ai/alameda/pkg/utils"
	datahubutilscontainer "github.com/containers-ai/alameda/pkg/utils/datahub/container"
	datahubutilspod "github.com/containers-ai/alameda/pkg/utils/datahub/pod"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	datahub_common "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"

	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

var (
	scope = logUtil.RegisterScope("operator_controllers", "operator controllers", 0)

	onceCheckHasOpenshiftAPIAppsV1 = sync.Once{}
	hasOpenshiftAPIAppsV1          = false

	requeueAfter = 3 * time.Second
)

var alamedascalerFirstSynced = false

// AlamedaScalerReconciler reconciles a AlamedaScaler object
type AlamedaScalerReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	ClusterUID string

	DatahubApplicationRepo *datahub_application.ApplicationRepository
	DatahubControllerRepo  *datahub_controller.ControllerRepository
	DatahubNamespaceRepo   *datahub_namespace.NamespaceRepository
	DatahubPodRepo         *datahub_pod.PodRepository
}

// Reconcile reads that state of the cluster for a AlamedaScaler object and makes changes based on the state read
// and what is in the AlamedaScaler .Spec
func (r *AlamedaScalerReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	if !alamedascalerFirstSynced {
		time.Sleep(5 * time.Second)
	}
	alamedascalerFirstSynced = true

	getResource := utilsresource.NewGetResource(r)
	listResources := utilsresource.NewListResources(r)
	updateResource := utilsresource.NewUpdateResource(r)

	onceCheckHasOpenshiftAPIAppsV1.Do(
		func() {
			exist, err := utils.ServerHasOpenshiftAPIAppsV1()
			if err != nil {
				panic(errors.Wrap(err, "Check if apiServer has openshift apps v1 api failed"))
			}
			hasOpenshiftAPIAppsV1 = exist
		})

	// Delete resources relative to AlamedaScaler
	if alamedaScaler, err := getResource.GetAlamedaScaler(req.Namespace, req.Name); err != nil && k8sErrors.IsNotFound(err) {
		scope.Infof("AlamedaScaler (%s/%s) is deleted, remove alameda pods from datahub.", req.Namespace, req.Name)
		if err := r.handleAlamedaScalerDeletion(req.Namespace, req.Name); err != nil {
			scope.Errorf("Handle AlamedaScaler(%s/%s) deletion failed, retry after %f seconds. %s", req.Namespace, req.Name, requeueAfter.Seconds(), err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: requeueAfter}, nil
		}
	} else if err == nil {
		alamedaScaler.SetDefaultValue()
		alamedaScalerNS := alamedaScaler.GetNamespace()
		alamedaScalerName := alamedaScaler.GetName()
		alamedascalerReconciler := alamedascaler_reconciler.NewReconciler(r, alamedaScaler)
		alamedascalerReconciler.ResetAlamedaController()

		scope.Infof(fmt.Sprintf("AlamedaScaler (%s/%s) found, try to sync latest alamedacontrollers.", alamedaScalerNS, alamedaScalerName))
		// select matched deployments
		if alamedaDeployments, err := listResources.ListDeploymentsByNamespaceLabels(req.Namespace, alamedaScaler.Spec.Selector.MatchLabels); err == nil {
			for _, alamedaDeployment := range alamedaDeployments {
				alamedaScaler, err = alamedascalerReconciler.UpdateStatusByDeployment(&alamedaDeployment)
				if err != nil {
					scope.Errorf("Update status of AlamedaScaler (%s/%s) by Deployment failed: %s", alamedaScalerNS, alamedaScalerName, err.Error())
					return ctrl.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
				}
			}
		} else {
			scope.Error(err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
		}

		// select matched deploymentConfigs
		if hasOpenshiftAPIAppsV1 {
			if alamedaDeploymentConfigs, err := listResources.ListDeploymentConfigsByNamespaceLabels(req.Namespace, alamedaScaler.Spec.Selector.MatchLabels); err == nil {
				for _, alamedaDeploymentConfig := range alamedaDeploymentConfigs {
					alamedaScaler, err = alamedascalerReconciler.UpdateStatusByDeploymentConfig(&alamedaDeploymentConfig)
					if err != nil {
						scope.Errorf("Update status of AlamedaScaler (%s/%s) by DeploymentConfig failed: %s", alamedaScalerNS, alamedaScalerName, err.Error())
						return ctrl.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
					}
				}
			} else {
				scope.Error(err.Error())
				return ctrl.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
			}
		}

		// select matched statefulSets
		if statefulSets, err := listResources.ListStatefulSetsByNamespaceLabels(req.Namespace, alamedaScaler.Spec.Selector.MatchLabels); err == nil {
			for _, statefulSet := range statefulSets {
				alamedaScaler, err = alamedascalerReconciler.UpdateStatusByStatefulSet(&statefulSet)
				if err != nil {
					scope.Errorf("update AlamedaScaler's (%s/%s) status by StatefulSets failed, retry reconciling: %s", req.Namespace, req.Name, err.Error())
					return ctrl.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
				}
			}
		} else {
			scope.Error(err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
		}

		if err := updateResource.UpdateAlamedaScaler(alamedaScaler); err != nil {
			scope.Errorf("Update AlamedaScaler (%s/%s) failed: %s", alamedaScalerNS, alamedaScalerName, err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
		}

		// after updating AlamedaPod in AlamedaScaler, start create AlamedaRecommendation if necessary and register alameda pod to datahub
		scope.Debugf("Start syncing AlamedaScaler (%s/%s) to datahub. %s", alamedaScalerNS, alamedaScalerName, alamutils.InterfaceToString(alamedaScaler))
		if err := r.syncAlamedaScalerWithDepResources(alamedaScaler); err != nil {
			scope.Error(err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
		}

	} else {
		scope.Errorf("get AlamedaScaler %s/%s failed: %s", req.Namespace, req.Name, err.Error())
		return ctrl.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
	}

	return ctrl.Result{}, nil
}

func (r *AlamedaScalerReconciler) syncAlamedaScalerWithDepResources(alamedaScaler *autoscalingv1alpha1.AlamedaScaler) error {

	existingPodsMap := make(map[autoscalingv1alpha1.NamespacedName]bool)
	existingPods := alamedaScaler.GetMonitoredPods()
	for _, pod := range existingPods {
		existingPodsMap[pod.GetNamespacedName()] = true
	}

	wg := errgroup.Group{}
	wg.Go(func() error {
		return r.syncDatahubResourceByAlamedaScaler(context.TODO(), *alamedaScaler)
	})
	wg.Go(func() error {
		return r.syncAlamedaRecommendation(alamedaScaler, existingPodsMap)
	})
	if err := wg.Wait(); err != nil {
		return errors.Wrapf(err, "sync AlamedaScaler %s/%s with dependent resources failed", alamedaScaler.Namespace, alamedaScaler.Name)
	}

	return nil
}

func (r *AlamedaScalerReconciler) syncDatahubResourceByAlamedaScaler(ctx context.Context, alamedaScaler autoscalingv1alpha1.AlamedaScaler) error {

	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		if err := r.syncDatahubApplicationsByAlamedaScaler(ctx, alamedaScaler); err != nil {
			return errors.Wrap(err, "sync applications with Datahub failed")
		}
		return nil
	})
	wg.Go(func() error {
		if err := r.syncDatahubControllersByAlamedaScaler(ctx, alamedaScaler); err != nil {
			return errors.Wrap(err, "sync controllers with Datahub failed")
		}
		return nil
	})
	wg.Go(func() error {
		if err := r.syncDatahubPodsByAlamedaScaler(ctx, alamedaScaler); err != nil {
			return errors.Wrap(err, "sync pods with Datahub failed")
		}
		return nil
	})
	return wg.Wait()
}

func (r *AlamedaScalerReconciler) syncDatahubApplicationsByAlamedaScaler(ctx context.Context, alamedaScaler autoscalingv1alpha1.AlamedaScaler) error {

	namespace := alamedaScaler.Namespace
	name := alamedaScaler.Name
	if err := r.DatahubApplicationRepo.CreateApplications([]*datahub_resources.Application{
		&datahub_resources.Application{
			ObjectMeta: &datahub_resources.ObjectMeta{
				Namespace:   namespace,
				Name:        name,
				ClusterName: r.ClusterUID,
			},
			AlamedaApplicationSpec: &datahub_resources.AlamedaApplicationSpec{
				ScalingTool: r.getAlamedaScalerDatahubScalingType(alamedaScaler),
			},
		},
	}); err != nil {
		return errors.Wrapf(err, "create Application(%s/%s) to Datahub failed", namespace, name)
	}
	return nil
}
func (r *AlamedaScalerReconciler) syncDatahubControllersByAlamedaScaler(ctx context.Context, alamedaScaler autoscalingv1alpha1.AlamedaScaler) error {

	namespace := alamedaScaler.Namespace
	name := alamedaScaler.Name
	if err := r.createAlamedaWatchedResourcesToDatahub(&alamedaScaler); err != nil {
		return errors.Wrapf(err, "create AlamedaScaler(%s/%s) watched resources to datahub failed", namespace, name)
	}

	// list all controller with namespace same as alamedaScaler
	controllers, err := r.listAlamedaWatchedResourcesToDatahub(&alamedaScaler)
	if err != nil {
		return errors.Wrapf(err, "list AlamedaScaler(%s/%s) watched resources to datahub failed", namespace, name)
	}
	err = r.deleteAlamedaWatchedResourcesToDatahub(context.TODO(), &alamedaScaler, controllers)
	if err != nil {
		return errors.Wrapf(err, "delete AlamedaScaler(%s/%s) watched resources to datahub failed", namespace, name)
	}
	return nil
}

func (r *AlamedaScalerReconciler) syncDatahubPodsByAlamedaScaler(ctx context.Context, alamedaScaler autoscalingv1alpha1.AlamedaScaler) error {

	// // When AlamedaScaler type is not vpa, delete all pods monitored by the AlamedaScaler from Datahub
	// if alamedaScaler.Spec.ScalingTool.Type != autoscalingv1alpha1.ScalingToolTypeVPA {
	// 	if err := r.deletePodsFromDatahubByAlamedaScaler(ctx, alamedaScaler.Namespace, alamedaScaler.Name); err != nil {
	// 		return errors.Wrap(err, "delete pods from Datahub by AlamedaScaler failed")
	// 	}
	// 	return nil
	// }

	// Create pods
	if err := r.createPodsToDatahubByAlamedaScaler(context.TODO(), alamedaScaler); err != nil {
		return errors.Wrapf(err, "create pods to Datahub by AlamedaScaler failed")
	}

	// Delete pods from Datahub which are no longer monitered by AlamedaScaler.
	monitoringPodMap := map[string]bool{}
	for _, pod := range alamedaScaler.GetMonitoredPods() {
		if pod == nil {
			continue
		}
		namespace := pod.Namespace
		name := pod.Name
		monitoringPodMap[fmt.Sprintf("%s/%s", namespace, name)] = true
	}
	podsNeedToBeDeleted := make([]*datahub_resources.ObjectMeta, 0)
	pods, err := r.DatahubPodRepo.ListAlamedaPodsByAlamedaScaler(context.TODO(), alamedaScaler.Namespace, alamedaScaler.Name)
	if err != nil {
		return errors.Wrapf(err, "list pods monitored by AlamedaScaler(%s/%s) from Datahub failed", alamedaScaler.Namespace, alamedaScaler.Name)
	}
	for _, pod := range pods {
		if pod == nil || pod.ObjectMeta == nil {
			continue
		}
		if ok, exist := monitoringPodMap[fmt.Sprintf("%s/%s", pod.ObjectMeta.Namespace, pod.ObjectMeta.Name)]; !ok || !exist {
			podsNeedToBeDeleted = append(podsNeedToBeDeleted, pod.ObjectMeta)
		}
	}
	if len(podsNeedToBeDeleted) > 0 {
		if r.DatahubPodRepo.DeletePods(context.TODO(), podsNeedToBeDeleted); err != nil {
			return errors.Wrap(err, "delete pods from Datahub failed")
		}
	}
	return nil
}

func (r *AlamedaScalerReconciler) listAlamedaWatchedResourcesToDatahub(scaler *autoscalingv1alpha1.AlamedaScaler) ([]*datahub_resources.Controller, error) {

	controllers, err := r.DatahubControllerRepo.ListControllersByApplication(context.TODO(), scaler.Namespace, scaler.Name)
	if err != nil {
		return nil, errors.Wrap(err, "list controllers by application from Datahub failed")
	}
	return controllers, nil
}

func (r *AlamedaScalerReconciler) createAlamedaWatchedResourcesToDatahub(scaler *autoscalingv1alpha1.AlamedaScaler) error {

	watchedReses := []*datahub_resources.Controller{}
	for _, dc := range scaler.Status.AlamedaController.DeploymentConfigs {
		policy := datahub_resources.RecommendationPolicy_RECOMMENDATION_POLICY_UNDEFINED
		if scaler.Spec.Policy == autoscalingv1alpha1.RecommendationPolicySTABLE {
			policy = datahub_resources.RecommendationPolicy_STABLE
		} else if scaler.Spec.Policy == autoscalingv1alpha1.RecommendationPolicyCOMPACT {
			policy = datahub_resources.RecommendationPolicy_COMPACT
		}
		watchedReses = append(watchedReses, &datahub_resources.Controller{
			ObjectMeta: &datahub_resources.ObjectMeta{
				Namespace:   dc.Namespace,
				Name:        dc.Name,
				ClusterName: r.ClusterUID,
			},
			Kind: datahub_resources.Kind_DEPLOYMENTCONFIG,
			AlamedaControllerSpec: &datahub_resources.AlamedaControllerSpec{
				AlamedaScaler: &datahub_resources.ObjectMeta{
					Namespace:   scaler.Namespace,
					Name:        scaler.Name,
					ClusterName: r.ClusterUID,
				},
				Policy:                        policy,
				EnableRecommendationExecution: scaler.IsEnableExecution(),
				ScalingTool:                   r.getAlamedaScalerDatahubScalingType(*scaler),
			},
			Replicas:     int32(len(dc.Pods)),
			SpecReplicas: *dc.SpecReplicas,
		})
	}
	for _, deploy := range scaler.Status.AlamedaController.Deployments {
		policy := datahub_resources.RecommendationPolicy_RECOMMENDATION_POLICY_UNDEFINED
		if scaler.Spec.Policy == autoscalingv1alpha1.RecommendationPolicySTABLE {
			policy = datahub_resources.RecommendationPolicy_STABLE
		} else if scaler.Spec.Policy == autoscalingv1alpha1.RecommendationPolicyCOMPACT {
			policy = datahub_resources.RecommendationPolicy_COMPACT
		}
		watchedReses = append(watchedReses, &datahub_resources.Controller{
			ObjectMeta: &datahub_resources.ObjectMeta{
				Namespace:   deploy.Namespace,
				Name:        deploy.Name,
				ClusterName: r.ClusterUID,
			},
			Kind: datahub_resources.Kind_DEPLOYMENT,
			AlamedaControllerSpec: &datahub_resources.AlamedaControllerSpec{
				AlamedaScaler: &datahub_resources.ObjectMeta{
					Namespace:   scaler.Namespace,
					Name:        scaler.Name,
					ClusterName: r.ClusterUID,
				},
				Policy:                        policy,
				EnableRecommendationExecution: scaler.IsEnableExecution(),
				ScalingTool:                   r.getAlamedaScalerDatahubScalingType(*scaler),
			},
			Replicas:     int32(len(deploy.Pods)),
			SpecReplicas: *deploy.SpecReplicas,
		})
	}
	for _, statefulSet := range scaler.Status.AlamedaController.StatefulSets {
		policy := datahub_resources.RecommendationPolicy_RECOMMENDATION_POLICY_UNDEFINED
		if scaler.Spec.Policy == autoscalingv1alpha1.RecommendationPolicySTABLE {
			policy = datahub_resources.RecommendationPolicy_STABLE
		} else if scaler.Spec.Policy == autoscalingv1alpha1.RecommendationPolicyCOMPACT {
			policy = datahub_resources.RecommendationPolicy_COMPACT
		}
		watchedReses = append(watchedReses, &datahub_resources.Controller{
			ObjectMeta: &datahub_resources.ObjectMeta{
				Namespace:   statefulSet.Namespace,
				Name:        statefulSet.Name,
				ClusterName: r.ClusterUID,
			},
			Kind: datahub_resources.Kind_STATEFULSET,
			AlamedaControllerSpec: &datahub_resources.AlamedaControllerSpec{
				AlamedaScaler: &datahub_resources.ObjectMeta{
					Namespace:   scaler.Namespace,
					Name:        scaler.Name,
					ClusterName: r.ClusterUID,
				},
				Policy:                        policy,
				EnableRecommendationExecution: scaler.IsEnableExecution(),
				ScalingTool:                   r.getAlamedaScalerDatahubScalingType(*scaler),
			},
			Replicas:     int32(len(statefulSet.Pods)),
			SpecReplicas: *statefulSet.SpecReplicas,
		})
	}
	err := r.DatahubControllerRepo.CreateControllers(watchedReses)
	return err
}

func (r *AlamedaScalerReconciler) deleteAlamedaWatchedResourcesToDatahub(ctx context.Context, scaler *autoscalingv1alpha1.AlamedaScaler, ctlrsFromDH []*datahub_resources.Controller) error {

	controllerMap := map[datahub_resources.Kind][]*datahub_resources.ObjectMeta{
		datahub_resources.Kind_DEPLOYMENT:       []*datahub_resources.ObjectMeta{},
		datahub_resources.Kind_DEPLOYMENTCONFIG: []*datahub_resources.ObjectMeta{},
		datahub_resources.Kind_STATEFULSET:      []*datahub_resources.ObjectMeta{},
	}
	for _, ctlr := range ctlrsFromDH {
		if !r.isControllerHasAlamedaScalerInfo(*ctlr, *scaler) {
			continue
		}
		if r.isControllerExistsInAlamedaScalerStatus(*ctlr, *scaler) {
			continue
		}
		controllerMap[ctlr.Kind] = append(controllerMap[ctlr.Kind], ctlr.ObjectMeta)
	}
	for kind, controllers := range controllerMap {
		if len(controllers) == 0 {
			continue
		}
		err := r.DatahubControllerRepo.DeleteControllers(ctx, controllers, kind)
		if err != nil {
			return errors.Wrap(err, "delete controllers from Datahub failed")
		}
	}
	return nil
}

func (r *AlamedaScalerReconciler) createPodsToDatahubByAlamedaScaler(ctx context.Context, scaler autoscalingv1alpha1.AlamedaScaler) error {

	pods := scaler.GetMonitoredPods()

	getResource := utilsresource.NewGetResource(r)

	policy := datahub_resources.RecommendationPolicy_STABLE
	if strings.ToLower(string(scaler.Spec.Policy)) == strings.ToLower(string(autoscalingv1alpha1.RecommendationPolicyCOMPACT)) {
		policy = datahub_resources.RecommendationPolicy_COMPACT
	} else if strings.ToLower(string(scaler.Spec.Policy)) == strings.ToLower(string(autoscalingv1alpha1.RecommendationPolicySTABLE)) {
		policy = datahub_resources.RecommendationPolicy_STABLE
	}

	podsNeedCreating := []*datahub_resources.Pod{}
	for _, pod := range pods {
		containers := []*datahub_resources.Container{}
		startTime := &timestamp.Timestamp{}
		for _, container := range pod.Containers {
			containers = append(containers, &datahub_resources.Container{
				Name: container.Name,
				Resources: &datahub_resources.ResourceRequirements{
					Limits:   map[int32]string{},
					Requests: map[int32]string{},
				},
			})
		}

		nodeName := ""
		resourceLink := ""
		podStatus := &datahub_resources.PodStatus{}
		replicas := int32(-1)
		if corePod, err := getResource.GetPod(pod.Namespace, pod.Name); err == nil {
			podStatus = datahubutilspod.NewStatus(corePod)
			replicas = datahubutilspod.GetReplicasFromPod(corePod, r)

			for _, containerStatus := range corePod.Status.ContainerStatuses {
				for containerIdx := range containers {
					if containerStatus.Name == containers[containerIdx].GetName() {
						containers[containerIdx].Status = datahubutilscontainer.NewStatus(&containerStatus)
						break
					}
				}
			}

			for _, podContainer := range corePod.Spec.Containers {
				for containerIdx := range containers {
					if podContainer.Name == containers[containerIdx].GetName() {
						for _, resourceType := range []corev1.ResourceName{
							corev1.ResourceCPU, corev1.ResourceMemory,
						} {
							if &podContainer.Resources != nil && podContainer.Resources.Limits != nil {
								resVal, ok := podContainer.Resources.Limits[resourceType]
								if ok && resourceType == corev1.ResourceCPU {
									containers[containerIdx].Resources.Limits[int32(datahub_common.ResourceName_CPU)] = strconv.FormatInt(resVal.MilliValue(), 10)
								}
								if ok && resourceType == corev1.ResourceMemory {
									containers[containerIdx].Resources.Limits[int32(datahub_common.ResourceName_MEMORY)] = strconv.FormatInt(resVal.Value(), 10)
								}
							}
							if &podContainer.Resources != nil && podContainer.Resources.Requests != nil {
								resVal, ok := podContainer.Resources.Requests[resourceType]
								if ok && resourceType == corev1.ResourceCPU {
									containers[containerIdx].Resources.Requests[int32(datahub_common.ResourceName_CPU)] = strconv.FormatInt(resVal.MilliValue(), 10)
								}
								if ok && resourceType == corev1.ResourceMemory {
									containers[containerIdx].Resources.Requests[int32(datahub_common.ResourceName_MEMORY)] = strconv.FormatInt(resVal.Value(), 10)
								}
							}
						}
						break
					}
				}
			}

			nodeName = corePod.Spec.NodeName
			startTime = &timestamp.Timestamp{
				Seconds: corePod.ObjectMeta.GetCreationTimestamp().Unix(),
			}
			resourceLink = utilsresource.GetResourceLinkForPod(r.Client, corePod)
			scope.Debugf(fmt.Sprintf("Resource link for pod (%s/%s) is %s", corePod.GetNamespace(), corePod.GetName(), resourceLink))
		} else {
			scope.Errorf("build Datahub pod to create failed, skip this pod: get pod %s/%s from k8s failed: %s", pod.Namespace, pod.Name, err.Error())
			continue
		}

		topCtrl, err := utils.ParseResourceLinkForTopController(resourceLink)

		if err != nil {
			scope.Error(err.Error())
		} else {
			topCtrl.Replicas = replicas
		}
		appName := fmt.Sprintf("%s-%s", scaler.Namespace, scaler.Name)
		if _, exist := scaler.Labels["app.federator.ai/name"]; exist {
			appName = scaler.Labels["app.federator.ai/name"]
		}
		appPartOf := appName
		if _, exist := scaler.Labels["app.federator.ai/part-of"]; exist {
			appPartOf = scaler.Labels["app.federator.ai/part-of"]
		}

		scalingTool := datahub_resources.ScalingTool_NONE
		scalingToolType := strings.ToLower(strings.Trim(scaler.Spec.ScalingTool.Type, " "))
		if scalingToolType == "vpa" {
			scalingTool = datahub_resources.ScalingTool_VPA
		} else if scalingToolType == "hpa" {
			scalingTool = datahub_resources.ScalingTool_HPA
		}

		podsNeedCreating = append(podsNeedCreating, &datahub_resources.Pod{
			AlamedaPodSpec: &datahub_resources.AlamedaPodSpec{
				AlamedaScaler: &datahub_resources.ObjectMeta{
					Namespace: scaler.Namespace,
					Name:      scaler.Name,
				},
				Policy:      datahub_resources.RecommendationPolicy(policy),
				ScalingTool: scalingTool,
				AlamedaScalerResources: &datahub_resources.ResourceRequirements{
					Requests: map[int32]string{
						int32(datahub_common.ResourceName_CPU):    scaler.GetRequestCPUMilliCores(),
						int32(datahub_common.ResourceName_MEMORY): scaler.GetRequestMemoryBytes(),
					},
					Limits: map[int32]string{
						int32(datahub_common.ResourceName_CPU):    scaler.GetLimitCPUMilliCores(),
						int32(datahub_common.ResourceName_MEMORY): scaler.GetLimitMemoryBytes(),
					},
				},
			},
			ObjectMeta: &datahub_resources.ObjectMeta{
				Name:        pod.Name,
				Namespace:   pod.Namespace,
				NodeName:    nodeName,
				ClusterName: r.ClusterUID,
			},
			Containers:    containers,
			ResourceLink:  resourceLink,
			StartTime:     startTime,
			TopController: topCtrl,
			Status:        podStatus,
			AppName:       appName,
			AppPartOf:     appPartOf,
		})
	}

	scope.Debugf("Create pods to datahub with request %s.", alamutils.InterfaceToString(podsNeedCreating))
	err := r.DatahubPodRepo.CreatePods(ctx, podsNeedCreating)
	if err != nil {
		return errors.Wrapf(err, "create pods monitored by AlamedaScaler(%s/%s) to Datahub failed", scaler.GetNamespace(), scaler.GetName())
	}
	scope.Infof(fmt.Sprintf("add alameda pods for AlamedaScaler (%s/%s) successfully", scaler.GetNamespace(), scaler.GetName()))

	return nil
}

func (r *AlamedaScalerReconciler) handleAlamedaScalerDeletion(namespace, name string) error {

	ctx := context.TODO()
	wg, ctx := errgroup.WithContext(ctx)
	wg.Go(func() error {
		if err := r.deleteControllersFromDatahubByAlamedaScaler(ctx, namespace, name); err != nil {
			return errors.Wrapf(err, "delete controllers from datahub by AlamedaScaler(%s/%s) failed", namespace, name)
		}
		scope.Infof("Remove alameda controllers of AlamedaScaler(%s/%s) from datahub successed.", namespace, name)
		return nil
	})
	wg.Go(func() error {
		if err := r.deletePodsFromDatahubByAlamedaScaler(ctx, namespace, name); err != nil {
			return errors.Wrapf(err, "delete pods from datahub by AlamedaScaler(%s/%s) failed", namespace, name)
		}
		scope.Infof("Remove alameda pods of AlamedaScaler(%s/%s) from datahub successed.", namespace, name)
		return nil
	})
	wg.Go(func() error {
		applicationObejctMeta := datahub_resources.ObjectMeta{
			Namespace:   namespace,
			Name:        name,
			ClusterName: r.ClusterUID,
		}
		// Delete application from datahub
		if err := r.DatahubApplicationRepo.DeleteApplications(ctx, []*datahub_resources.ObjectMeta{&applicationObejctMeta}); err != nil {
			return errors.Wrapf(err, "delete Application(%s/%s) from datahub failed", namespace, name)
		} else {
			if r.DatahubNamespaceRepo.IsNSExcluded(namespace) {
				r.DatahubNamespaceRepo.DeleteNamespaces([]*datahub_resources.Namespace{
					&datahub_resources.Namespace{
						ObjectMeta: &datahub_resources.ObjectMeta{
							Name:        namespace,
							ClusterName: r.ClusterUID,
						},
					},
				})
			}
		}
		scope.Infof("Remove application of AlamedaScaler(%s/%s) from datahub successed.", namespace, name)
		return nil
	})

	return wg.Wait()
}

func (r *AlamedaScalerReconciler) deleteControllersFromDatahubByAlamedaScaler(ctx context.Context, namespace, name string) error {

	application, err := r.DatahubApplicationRepo.GetApplication(ctx, namespace, name)
	if err != nil {
		return errors.Wrap(err, "get application failed")
	}

	controllerMap := map[datahub_resources.Kind][]*datahub_resources.ObjectMeta{
		datahub_resources.Kind_DEPLOYMENT:       []*datahub_resources.ObjectMeta{},
		datahub_resources.Kind_DEPLOYMENTCONFIG: []*datahub_resources.ObjectMeta{},
		datahub_resources.Kind_STATEFULSET:      []*datahub_resources.ObjectMeta{},
	}
	for _, controller := range application.Controllers {
		controllerMap[controller.Kind] = append(controllerMap[controller.Kind], controller.ObjectMeta)
	}
	wg, wgCTX := errgroup.WithContext(ctx)
	for kind, objectMetas := range controllerMap {
		if len(objectMetas) == 0 {
			continue
		}
		copyKind := kind
		copyObjectMetas := objectMetas
		wg.Go(func() error {
			return r.DatahubControllerRepo.DeleteControllers(wgCTX, copyObjectMetas, copyKind)
		})
	}

	return wg.Wait()
}

func (r *AlamedaScalerReconciler) deletePodsFromDatahubByAlamedaScaler(ctx context.Context, namespace, name string) error {

	pods, err := r.DatahubPodRepo.ListAlamedaPodsByAlamedaScaler(ctx, namespace, name)
	if err != nil {
		return errors.Wrapf(err, "list pods by AlamedaScaler(%s/%s) failed", namespace, name)
	}

	podsNeedDeleting := make([]*datahub_resources.ObjectMeta, len(pods))
	for i, pod := range pods {
		podsNeedDeleting[i] = pod.ObjectMeta
	}
	if len(podsNeedDeleting) == 0 {
		return nil
	}
	err = r.DatahubPodRepo.DeletePods(ctx, podsNeedDeleting)
	if err != nil {
		return errors.Wrapf(err, "delete pods by AlamedaScaler(%s/%s) failed", namespace, name)
	}
	return nil
}

func (r *AlamedaScalerReconciler) syncAlamedaRecommendation(alamedaScaler *autoscalingv1alpha1.AlamedaScaler, existingPodsMap map[autoscalingv1alpha1.NamespacedName]bool) error {

	currentPods := alamedaScaler.GetMonitoredPods()

	if err := r.createAssociateRecommendation(alamedaScaler, currentPods); err != nil {
		return errors.Wrapf(err, "sync AlamedaRecommendation failed: %s", err.Error())
	}

	if err := r.deleteAlamedaRecommendations(alamedaScaler, existingPodsMap); err != nil {
		return errors.Wrapf(err, "sync AlamedaRecommendation failed: %s", err.Error())
	}

	return nil
}

func (r *AlamedaScalerReconciler) createAssociateRecommendation(alamedaScaler *autoscalingv1alpha1.AlamedaScaler, pods []*autoscalingv1alpha1.AlamedaPod) error {

	getResource := utilsresource.NewGetResource(r)
	m := alamedaScaler.GetLabelMapToSetToAlamedaRecommendationLabel()

	for _, pod := range pods {

		// try to create the recommendation by pod
		recommendationNS := pod.Namespace
		recommendationName := pod.Name

		recommendation := &autoscalingv1alpha1.AlamedaRecommendation{
			ObjectMeta: metav1.ObjectMeta{
				Name:      recommendationName,
				Namespace: recommendationNS,
				Labels:    m,
			},
			Spec: autoscalingv1alpha1.AlamedaRecommendationSpec{
				Containers: pod.Containers,
			},
		}

		err := controllerutil.SetControllerReference(alamedaScaler, recommendation, r.Scheme)
		if err != nil {
			scope.Errorf("set Recommendation %s/%s ownerReference failed, skip create Recommendation to kubernetes, error message: %s", alamedaScaler.Namespace, alamedaScaler.Name, err.Error())
			continue
		}
		_, err = getResource.GetAlamedaRecommendation(recommendationNS, recommendationName)
		if err != nil && k8sErrors.IsNotFound(err) {
			err = r.Create(context.TODO(), recommendation)
			if err != nil {
				return errors.Wrapf(err, "create recommendation %s/%s to kuernetes failed: %s", alamedaScaler.Namespace, alamedaScaler.Name, err.Error())
			}
		}
	}
	return nil
}

func (r *AlamedaScalerReconciler) listAlamedaRecommendationsOwnedByAlamedaScaler(alamedaScaler *autoscalingv1alpha1.AlamedaScaler) ([]*autoscalingv1alpha1.AlamedaRecommendation, error) {

	listResource := utilsresource.NewListResources(r)
	tmp := make([]*autoscalingv1alpha1.AlamedaRecommendation, 0)

	alamedaRecommendations, err := listResource.ListAlamedaRecommendationOwnedByAlamedaScaler(alamedaScaler)
	if err != nil {
		return tmp, err
	}

	for _, alamedaRecommendation := range alamedaRecommendations {
		cpAlamedaRecommendation := alamedaRecommendation
		tmp = append(tmp, &cpAlamedaRecommendation)
	}

	return tmp, nil
}

func (r *AlamedaScalerReconciler) deleteAlamedaRecommendations(alamedaScaler *autoscalingv1alpha1.AlamedaScaler, existingPodsMap map[autoscalingv1alpha1.NamespacedName]bool) error {

	alamedaRecommendations, err := r.getNeedDeletingAlamedaRecommendations(alamedaScaler, existingPodsMap)
	if err != nil {
		return errors.Wrapf(err, "delete AlamedaRecommendations failed: %s", err.Error())
	}

	for _, alamedaRecommendation := range alamedaRecommendations {

		recommendationNS := alamedaRecommendation.Namespace
		recommendationName := alamedaRecommendation.Name

		recommendation := &autoscalingv1alpha1.AlamedaRecommendation{
			ObjectMeta: metav1.ObjectMeta{
				Name:      recommendationName,
				Namespace: recommendationNS,
			},
		}

		if err := r.Delete(context.TODO(), recommendation); err != nil {
			return errors.Wrapf(err, "delete AlamedaRecommendations %s/%s to kuernetes failed: %s", recommendationNS, recommendationName, err.Error())
		}
	}

	return nil
}

func (r *AlamedaScalerReconciler) getNeedDeletingAlamedaRecommendations(alamedaScaler *autoscalingv1alpha1.AlamedaScaler, existingPodsMap map[autoscalingv1alpha1.NamespacedName]bool) ([]*autoscalingv1alpha1.AlamedaRecommendation, error) {

	needDeletingAlamedaRecommendations := make([]*autoscalingv1alpha1.AlamedaRecommendation, 0)
	alamedaRecommendations, err := r.listAlamedaRecommendationsOwnedByAlamedaScaler(alamedaScaler)
	if err != nil {
		return needDeletingAlamedaRecommendations, errors.Wrapf(err, "get need deleting AlamedaRecommendations failed: %s", err.Error())
	}
	for _, alamedaRecommendation := range alamedaRecommendations {
		cpAlamedaRecommendation := *alamedaRecommendation
		namespacedName := alamedaRecommendation.GetNamespacedName()
		if isExisting, exist := existingPodsMap[namespacedName]; !exist || !isExisting {
			needDeletingAlamedaRecommendations = append(needDeletingAlamedaRecommendations, &cpAlamedaRecommendation)
		}
	}

	return needDeletingAlamedaRecommendations, nil
}

func (r *AlamedaScalerReconciler) getAlamedaScalerDatahubScalingType(alamedaScaler autoscalingv1alpha1.AlamedaScaler) datahub_resources.ScalingTool {
	scalingType := datahub_resources.ScalingTool_SCALING_TOOL_UNDEFINED
	switch alamedaScaler.Spec.ScalingTool.Type {
	case autoscalingv1alpha1.ScalingToolTypeVPA:
		scalingType = datahub_resources.ScalingTool_VPA
	case autoscalingv1alpha1.ScalingToolTypeHPA:
		scalingType = datahub_resources.ScalingTool_HPA
	case autoscalingv1alpha1.ScalingToolTypeDefault:
		scalingType = datahub_resources.ScalingTool_NONE
	}
	return scalingType
}

func (r *AlamedaScalerReconciler) isControllerHasAlamedaScalerInfo(controller datahub_resources.Controller, alamedaScaler autoscalingv1alpha1.AlamedaScaler) bool {
	if controller.AlamedaControllerSpec == nil || controller.AlamedaControllerSpec.AlamedaScaler == nil {
		return false
	}
	if controller.AlamedaControllerSpec.AlamedaScaler.Namespace == alamedaScaler.Namespace && controller.AlamedaControllerSpec.AlamedaScaler.Name == alamedaScaler.Name {
		return true
	}
	return false
}

func (r *AlamedaScalerReconciler) isControllerExistsInAlamedaScalerStatus(controller datahub_resources.Controller, alamedaScaler autoscalingv1alpha1.AlamedaScaler) bool {

	isInAlamedaScaler := false
	switch controller.Kind {
	case datahub_resources.Kind_DEPLOYMENTCONFIG:
		for _, dc := range alamedaScaler.Status.AlamedaController.DeploymentConfigs {
			if controller.ObjectMeta.Name == dc.Name {
				isInAlamedaScaler = true
				break
			}
		}
	case datahub_resources.Kind_DEPLOYMENT:
		for _, deploy := range alamedaScaler.Status.AlamedaController.Deployments {
			if controller.ObjectMeta.Name == deploy.Name {
				isInAlamedaScaler = true
				break
			}
		}
	case datahub_resources.Kind_STATEFULSET:
		for _, statefulSet := range alamedaScaler.Status.AlamedaController.StatefulSets {
			if controller.ObjectMeta.Name == statefulSet.Name {
				isInAlamedaScaler = true
				break
			}
		}
	}
	return isInAlamedaScaler
}

func (r *AlamedaScalerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&autoscalingv1alpha1.AlamedaScaler{}).
		Complete(r)
}
