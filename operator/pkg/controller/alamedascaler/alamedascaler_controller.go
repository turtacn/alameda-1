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

package alamedascaler

import (
	"context"
	"fmt"
	"golang.org/x/sync/errgroup"
	"strconv"
	"strings"
	"sync"
	"time"

	datahubclient "github.com/containers-ai/alameda/operator/datahub/client"
	datahub_application "github.com/containers-ai/alameda/operator/datahub/client/application"
	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/pkg/apis/autoscaling/v1alpha1"
	alamedascaler_reconciler "github.com/containers-ai/alameda/operator/pkg/reconciler/alamedascaler"
	"github.com/containers-ai/alameda/operator/pkg/utils"
	datahubutils "github.com/containers-ai/alameda/operator/pkg/utils/datahub"
	utilsresource "github.com/containers-ai/alameda/operator/pkg/utils/resources"
	alamutils "github.com/containers-ai/alameda/pkg/utils"
	datahubutilscontainer "github.com/containers-ai/alameda/pkg/utils/datahub/container"
	datahubutilspod "github.com/containers-ai/alameda/pkg/utils/datahub/pod"
	k8sutils "github.com/containers-ai/alameda/pkg/utils/kubernetes"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_common "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	scope = logUtil.RegisterScope("alamedascaler", "alamedascaler log", 0)

	onceCheckHasOpenshiftAPIAppsV1 = sync.Once{}
	hasOpenshiftAPIAppsV1          = false
	grpcDefaultRetry               = uint(3)
)

var cachedFirstSynced = false

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AlamedaScaler Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this autoscaling.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	conn, _ := grpc.Dial(datahubutils.GetDatahubAddress(), grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(grpc_retry.WithMax(grpcDefaultRetry))))

	k8sClient, err := client.New(mgr.GetConfig(), client.Options{})
	if err != nil {
		panic(errors.Wrap(err, "new kuberenetes client failed").Error())
	}
	clusterUID, err := k8sutils.GetClusterUID(k8sClient)
	if err != nil || clusterUID == "" {
		panic("cannot get cluster uid")
	}

	datahubApplicationRepo := datahub_application.NewApplicationRepository(conn, clusterUID)

	return &ReconcileAlamedaScaler{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),

		clusterUID: clusterUID,

		datahubApplicationRepo: datahubApplicationRepo,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("alamedascaler-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		scope.Error(err.Error())
		return err
	}

	if err = c.Watch(&source.Kind{Type: &autoscalingv1alpha1.AlamedaScaler{}}, &handler.EnqueueRequestForObject{}); err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAlamedaScaler{}

// ReconcileAlamedaScaler reconciles a AlamedaScaler object
type ReconcileAlamedaScaler struct {
	client.Client
	scheme *runtime.Scheme

	clusterUID string

	datahubApplicationRepo *datahub_application.ApplicationRepository
}

// Reconcile reads that state of the cluster for a AlamedaScaler object and makes changes based on the state read
// and what is in the AlamedaScaler .Spec
func (r *ReconcileAlamedaScaler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	if !cachedFirstSynced {
		time.Sleep(5 * time.Second)
	}
	cachedFirstSynced = true

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

	// Take care of AlamedaScaler
	if alamedaScaler, err := getResource.GetAlamedaScaler(request.Namespace, request.Name); err != nil && k8sErrors.IsNotFound(err) {
		scope.Infof("AlamedaScaler (%s/%s) is deleted, remove alameda pods from datahub.", request.Namespace, request.Name)
		err := r.deletePodsFromDatahub(&request.NamespacedName, make(map[autoscalingv1alpha1.NamespacedName]bool))
		if err != nil {
			scope.Errorf("Remove alameda pods of alamedascaler (%s/%s) from datahub failed. %s", request.Namespace, request.Name, err.Error())
		} else {
			scope.Infof("Remove alameda pods of alamedascaler (%s/%s) from datahub successed.", request.Namespace, request.Name)
		}
		err = r.deleteControllersFromDatahub(request.Namespace, request.Name)
		if err != nil {
			scope.Errorf("Remove alameda controllers of alamedascaler (%s/%s) from datahub failed. %s", request.Namespace, request.Name, err.Error())
		} else {
			scope.Infof("Remove alameda controllers of alamedascaler (%s/%s) from datahub successed.", request.Namespace, request.Name)
		}

		// Delete application from datahub
		err = r.datahubApplicationRepo.DeleteApplications([]*datahub_resources.Application{
			&datahub_resources.Application{
				ObjectMeta: &datahub_resources.ObjectMeta{
					Name:        request.NamespacedName.Name,
					Namespace:   request.NamespacedName.Namespace,
					ClusterName: r.clusterUID,
				},
			},
		})
		if err != nil {
			scope.Errorf("Delete application %s/%s from datahub failed: %s",
				request.NamespacedName.Namespace, request.NamespacedName.Name, err.Error())
		}
	} else if err == nil {
		// TODO: deployment already in the AlamedaScaler cannot join the other
		alamedaScaler.SetDefaultValue()
		alamedaScalerNS := alamedaScaler.GetNamespace()
		alamedaScalerName := alamedaScaler.GetName()
		alamedascalerReconciler := alamedascaler_reconciler.NewReconciler(r, alamedaScaler)
		alamedascalerReconciler.ResetAlamedaController()

		scope.Infof(fmt.Sprintf("AlamedaScaler (%s/%s) found, try to sync latest alamedacontrollers.", alamedaScalerNS, alamedaScalerName))
		// select matched deployments
		if alamedaDeployments, err := listResources.ListDeploymentsByNamespaceLabels(request.Namespace, alamedaScaler.Spec.Selector.MatchLabels); err == nil {
			for _, alamedaDeployment := range alamedaDeployments {
				alamedaScaler, err = alamedascalerReconciler.UpdateStatusByDeployment(&alamedaDeployment)
				if err != nil {
					scope.Errorf("Update status of AlamedaScaler (%s/%s) by Deployment failed: %s", alamedaScalerNS, alamedaScalerName, err.Error())
					return reconcile.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
				}
			}
		} else {
			scope.Error(err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
		}

		// select matched deploymentConfigs
		if hasOpenshiftAPIAppsV1 {
			if alamedaDeploymentConfigs, err := listResources.ListDeploymentConfigsByNamespaceLabels(request.Namespace, alamedaScaler.Spec.Selector.MatchLabels); err == nil {
				for _, alamedaDeploymentConfig := range alamedaDeploymentConfigs {
					alamedaScaler, err = alamedascalerReconciler.UpdateStatusByDeploymentConfig(&alamedaDeploymentConfig)
					if err != nil {
						scope.Errorf("Update status of AlamedaScaler (%s/%s) by DeploymentConfig failed: %s", alamedaScalerNS, alamedaScalerName, err.Error())
						return reconcile.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
					}
				}
			} else {
				scope.Error(err.Error())
				return reconcile.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
			}
		}

		// select matched statefulSets
		if statefulSets, err := listResources.ListStatefulSetsByNamespaceLabels(request.Namespace, alamedaScaler.Spec.Selector.MatchLabels); err == nil {
			for _, statefulSet := range statefulSets {
				alamedaScaler, err = alamedascalerReconciler.UpdateStatusByStatefulSet(&statefulSet)
				if err != nil {
					scope.Errorf("update AlamedaScaler's (%s/%s) status by StatefulSets failed, retry reconciling: %s", request.Namespace, request.Name, err.Error())
					return reconcile.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
				}
			}
		} else {
			scope.Error(err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
		}

		if err := updateResource.UpdateAlamedaScaler(alamedaScaler); err != nil {
			scope.Errorf("Update AlamedaScaler (%s/%s) failed: %s", alamedaScalerNS, alamedaScalerName, err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
		}

		if err := r.createAlamedaWatchedResourcesToDatahub(alamedaScaler); err != nil {
			scope.Errorf("Create AlamedaScaler (%s/%s) watched resources to datahub failed: %s", alamedaScalerNS, alamedaScalerName, err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
		}

		// list all controller with namespace same as alamedaScaler
		controllers, err := r.listAlamedaWatchedResourcesToDatahub(alamedaScaler)
		if err != nil {
			scope.Errorf("List AlamedaScaler (%s/%s) watched resources to datahub failed: %s", alamedaScalerNS, alamedaScalerName, err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
		}

		err = r.deleteAlamedaWatchedResourcesToDatahub(alamedaScaler, controllers)
		if err != nil {
			scope.Errorf("Delete AlamedaScaler (%s/%s) watched resources to datahub failed: %s", alamedaScalerNS, alamedaScalerName, err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
		}

		// after updating AlamedaPod in AlamedaScaler, start create AlamedaRecommendation if necessary and register alameda pod to datahub
		scope.Debugf("Start syncing AlamedaScaler (%s/%s) to datahub. %s", alamedaScalerNS, alamedaScalerName, alamutils.InterfaceToString(alamedaScaler))
		if err := r.syncAlamedaScalerWithDepResources(alamedaScaler); err != nil {
			scope.Error(err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
		}

		// add application from datahub
		err = r.datahubApplicationRepo.CreateApplications([]*datahub_resources.Application{
			&datahub_resources.Application{
				ObjectMeta: &datahub_resources.ObjectMeta{
					Name:        request.NamespacedName.Name,
					Namespace:   request.NamespacedName.Namespace,
					ClusterName: r.clusterUID,
				},
				AlamedaApplicationSpec: &datahub_resources.AlamedaApplicationSpec{
					ScalingTool: r.getAlamedaScalerDatahubScalingType(*alamedaScaler),
				},
			},
		})
		if err != nil {
			scope.Errorf("Create application %s/%s from datahub failed: %s",
				request.NamespacedName.Namespace, request.NamespacedName.Name, err.Error())
		}
	} else {
		scope.Errorf("get AlamedaScaler %s/%s failed: %s", request.Namespace, request.Name, err.Error())
		return reconcile.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileAlamedaScaler) syncAlamedaScalerWithDepResources(alamedaScaler *autoscalingv1alpha1.AlamedaScaler) error {

	existingPodsMap := make(map[autoscalingv1alpha1.NamespacedName]bool)
	existingPods := alamedaScaler.GetMonitoredPods()
	for _, pod := range existingPods {
		existingPodsMap[pod.GetNamespacedName()] = true
	}

	wg := errgroup.Group{}
	wg.Go(func() error {
		return r.syncDatahubResource(alamedaScaler, existingPodsMap)
	})
	wg.Go(func() error {
		return r.syncAlamedaRecommendation(alamedaScaler, existingPodsMap)
	})
	if err := wg.Wait(); err != nil {
		return errors.Wrapf(err, "sync AlamedaScaler %s/%s with dependent resources failed", alamedaScaler.Namespace, alamedaScaler.Name)
	}

	return nil
}

func (r *ReconcileAlamedaScaler) syncDatahubResource(alamedaScaler *autoscalingv1alpha1.AlamedaScaler, existingPodsMap map[autoscalingv1alpha1.NamespacedName]bool) error {

	currentPods := alamedaScaler.GetMonitoredPods()

	if len(currentPods) > 0 {
		if err := r.createPodsToDatahub(alamedaScaler, currentPods); err != nil {
			return errors.Wrapf(err, "sync Datahub resource failed: %s", err.Error())
		}
	}

	if err := r.deletePodsFromDatahub(&types.NamespacedName{
		Namespace: alamedaScaler.GetNamespace(),
		Name:      alamedaScaler.GetName(),
	}, existingPodsMap); err != nil {
		return errors.Wrapf(err, "sync Datahub resource failed: %s", err.Error())
	}

	return nil
}

func (r *ReconcileAlamedaScaler) listAlamedaWatchedResourcesToDatahub(scaler *autoscalingv1alpha1.AlamedaScaler) ([]*datahub_resources.Controller, error) {
	k8sRes := datahubclient.NewK8SResource(r.clusterUID)
	controllers, err := k8sRes.ListAlamedaWatchedResource(scaler.GetNamespace(), "")
	return controllers, err
}

func (r *ReconcileAlamedaScaler) createAlamedaWatchedResourcesToDatahub(scaler *autoscalingv1alpha1.AlamedaScaler) error {
	k8sRes := datahubclient.NewK8SResource(r.clusterUID)
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
				ClusterName: r.clusterUID,
			},
			Kind: datahub_resources.Kind_DEPLOYMENTCONFIG,
			AlamedaControllerSpec: &datahub_resources.AlamedaControllerSpec{
				AlamedaScaler: &datahub_resources.ObjectMeta{
					Namespace:   scaler.Namespace,
					Name:        scaler.Name,
					ClusterName: r.clusterUID,
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
				ClusterName: r.clusterUID,
			},
			Kind: datahub_resources.Kind_DEPLOYMENT,
			AlamedaControllerSpec: &datahub_resources.AlamedaControllerSpec{
				AlamedaScaler: &datahub_resources.ObjectMeta{
					Namespace:   scaler.Namespace,
					Name:        scaler.Name,
					ClusterName: r.clusterUID,
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
				ClusterName: r.clusterUID,
			},
			Kind: datahub_resources.Kind_STATEFULSET,
			AlamedaControllerSpec: &datahub_resources.AlamedaControllerSpec{
				AlamedaScaler: &datahub_resources.ObjectMeta{
					Namespace:   scaler.Namespace,
					Name:        scaler.Name,
					ClusterName: r.clusterUID,
				},
				Policy:                        policy,
				EnableRecommendationExecution: scaler.IsEnableExecution(),
				ScalingTool:                   r.getAlamedaScalerDatahubScalingType(*scaler),
			},
			Replicas:     int32(len(statefulSet.Pods)),
			SpecReplicas: *statefulSet.SpecReplicas,
		})
	}
	err := k8sRes.CreateAlamedaWatchedResource(watchedReses)
	return err
}

func (r *ReconcileAlamedaScaler) deleteAlamedaWatchedResourcesToDatahub(scaler *autoscalingv1alpha1.AlamedaScaler, ctlrsFromDH []*datahub_resources.Controller) error {
	delCtlrs := []*datahub_resources.Controller{}

	for _, ctlr := range ctlrsFromDH {
		copyCTRL := *ctlr
		ctScalerNS := ctlr.GetAlamedaControllerSpec().GetAlamedaScaler().GetNamespace()
		ctScalerName := ctlr.GetAlamedaControllerSpec().GetAlamedaScaler().GetName()
		if ctScalerNS == scaler.GetNamespace() && ctScalerName == scaler.GetName() {
			continue
		}

		ctlrKind := ctlr.GetKind()
		ctlrName := ctlr.GetObjectMeta().GetName()
		inScaler := false
		if ctlrKind == datahub_resources.Kind_DEPLOYMENTCONFIG {
			for _, dc := range scaler.Status.AlamedaController.DeploymentConfigs {
				if ctlrName == dc.Name {
					inScaler = true
					break
				}
			}
			if !inScaler {
				delCtlrs = append(delCtlrs, &copyCTRL)
			}
		} else if ctlrKind == datahub_resources.Kind_DEPLOYMENT {
			for _, deploy := range scaler.Status.AlamedaController.Deployments {
				if ctlrName == deploy.Name {
					inScaler = true
					break
				}
			}
			if !inScaler {
				delCtlrs = append(delCtlrs, &copyCTRL)
			}
		} else if ctlrKind == datahub_resources.Kind_STATEFULSET {
			for _, statefulSet := range scaler.Status.AlamedaController.StatefulSets {
				if ctlrName == statefulSet.Name {
					inScaler = true
					break
				}
			}
			if !inScaler {
				delCtlrs = append(delCtlrs, &copyCTRL)
			}
		}
	}

	k8sRes := datahubclient.NewK8SResource(r.clusterUID)
	if len(delCtlrs) > 0 {
		err := k8sRes.DeleteAlamedaWatchedResource(delCtlrs)
		return err
	}
	return nil
}

func (r *ReconcileAlamedaScaler) createPodsToDatahub(scaler *autoscalingv1alpha1.AlamedaScaler, pods []*autoscalingv1alpha1.AlamedaPod) error {

	getResource := utilsresource.NewGetResource(r)

	conn, err := grpc.Dial(datahubutils.GetDatahubAddress(), grpc.WithInsecure())
	if err != nil {
		return errors.Errorf("create pods to datahub failed: %s", err.Error())
	}

	defer conn.Close()
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(conn)

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
				ClusterName: r.clusterUID,
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

	req := datahub_resources.CreatePodsRequest{
		Pods: podsNeedCreating,
	}
	scope.Debugf("Create pods to datahub with request %s.", alamutils.InterfaceToString(req))
	resp, err := datahubServiceClnt.CreatePods(context.Background(), &req)
	if err != nil {
		return errors.Errorf("add alameda pods for AlamedaScaler (%s/%s) failed: %s", scaler.GetNamespace(), scaler.GetName(), err.Error())
	} else if resp.Code != int32(code.Code_OK) {
		return errors.Errorf("add alameda pods for AlamedaScaler (%s/%s) failed: receive response: code: %d, message: %s", scaler.GetNamespace(), scaler.GetName(), resp.Code, resp.Message)
	}
	scope.Infof(fmt.Sprintf("add alameda pods for AlamedaScaler (%s/%s) successfully", scaler.GetNamespace(), scaler.GetName()))

	return nil
}

func (r *ReconcileAlamedaScaler) deleteControllersFromDatahub(scalerNamespace, scalerName string) error {

	k8sRes := datahubclient.NewK8SResource(r.clusterUID)
	controllers, err := k8sRes.ListAlamedaWatchedResource(scalerNamespace, "")
	if err != nil {
		return err
	}

	controllersNeedDelete := make([]*datahub_resources.Controller, 0)
	for _, controller := range controllers {
		ctScalerNS := controller.GetAlamedaControllerSpec().GetAlamedaScaler().GetNamespace()
		ctScalerName := controller.GetAlamedaControllerSpec().GetAlamedaScaler().GetName()
		if ctScalerNS == scalerNamespace && ctScalerName == scalerName {
			controllersNeedDelete = append(controllersNeedDelete, controller)
		}
	}

	return k8sRes.DeleteAlamedaWatchedResource(controllersNeedDelete)
}

func (r *ReconcileAlamedaScaler) deletePodsFromDatahub(scalerNamespacedName *types.NamespacedName, existingPodsMap map[autoscalingv1alpha1.NamespacedName]bool) error {

	pods, err := r.getPodsNeedDeleting(scalerNamespacedName, existingPodsMap)
	if err != nil {
		return errors.Wrapf(err, "delete pods from datahub failed: %s", err.Error())
	}

	conn, err := grpc.Dial(datahubutils.GetDatahubAddress(), grpc.WithInsecure())
	if err != nil {
		return errors.Errorf("delete pods from datahub failed: %s", err.Error())
	}

	defer conn.Close()
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(conn)

	podsNeedDeleting := []*datahub_resources.ObjectMeta{}
	for _, pod := range pods {
		podsNeedDeleting = append(podsNeedDeleting, &datahub_resources.ObjectMeta{
			Namespace:   pod.Namespace,
			Name:        pod.Name,
			ClusterName: r.clusterUID,
		})
	}

	req := datahub_resources.DeletePodsRequest{
		ObjectMeta: podsNeedDeleting,
	}
	resp, err := datahubServiceClnt.DeletePods(context.Background(), &req)
	if err != nil {
		return errors.Errorf("remove alameda pods for AlamedaScaler (%s/%s) failed: %s", scalerNamespacedName.Namespace, scalerNamespacedName.Name, err.Error())
	} else if resp.Code != int32(code.Code_OK) {
		return errors.Errorf("remove alameda pods for AlamedaScaler (%s/%s) failed: receive response: code: %d, message: %s", scalerNamespacedName.Namespace, scalerNamespacedName.Name, resp.Code, resp.Message)
	}
	scope.Infof(fmt.Sprintf("remove alameda pods for AlamedaScaler (%s/%s) successfully", scalerNamespacedName.Namespace, scalerNamespacedName.Name))

	return nil
}

func (r *ReconcileAlamedaScaler) getPodsNeedDeleting(scalerNamespacedName *types.NamespacedName, existingPodsMap map[autoscalingv1alpha1.NamespacedName]bool) ([]*autoscalingv1alpha1.AlamedaPod, error) {

	copyScaler := *scalerNamespacedName

	needDeletingPods := make([]*autoscalingv1alpha1.AlamedaPod, 0)
	podsInDatahub, err := r.getPodsObservedByAlamedaScalerFromDatahub(&copyScaler)
	if err != nil {
		return needDeletingPods, errors.Wrapf(err, "get pods need deleting failed: %s", err.Error())
	}
	for _, pod := range podsInDatahub {
		namespacedName := pod.GetNamespacedName()
		if isExisting, exist := existingPodsMap[namespacedName]; !exist || !isExisting {
			needDeletingPods = append(needDeletingPods, &autoscalingv1alpha1.AlamedaPod{
				Namespace: pod.Namespace,
				Name:      pod.Name,
			})
		}
	}

	return needDeletingPods, nil
}

func (r *ReconcileAlamedaScaler) getPodsObservedByAlamedaScalerFromDatahub(scalerNamespacedName *types.NamespacedName) ([]*autoscalingv1alpha1.AlamedaPod, error) {

	podsInDatahub := make([]*autoscalingv1alpha1.AlamedaPod, 0)

	conn, err := grpc.Dial(datahubutils.GetDatahubAddress(), grpc.WithInsecure())
	if err != nil {
		return podsInDatahub, errors.Errorf("get pods from datahub failed: %s", err.Error())
	}

	defer conn.Close()
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(conn)

	req := datahub_resources.ListPodsRequest{
		ObjectMeta: []*datahub_resources.ObjectMeta{
			&datahub_resources.ObjectMeta{
				Namespace:   scalerNamespacedName.Namespace,
				Name:        scalerNamespacedName.Name,
				ClusterName: r.clusterUID,
			},
		},
		Kind: datahub_resources.Kind_ALAMEDASCALER,
	}
	resp, err := datahubServiceClnt.ListPods(context.Background(), &req)
	if err != nil {
		return podsInDatahub, errors.Errorf("get alameda pods for AlamedaScaler (%s/%s) failed: %s", scalerNamespacedName.Namespace, scalerNamespacedName.Name, err.Error())
	} else if resp.Status == nil {
		return podsInDatahub, errors.Errorf("get alameda pods for AlamedaScaler (%s/%s) failed: receive null status", scalerNamespacedName.Namespace, scalerNamespacedName.Name)
	} else if resp.Status.Code != int32(code.Code_OK) {
		return podsInDatahub, errors.Errorf("get alameda pods for AlamedaScaler (%s/%s) failed: receive response: code: %d, message: %s", scalerNamespacedName.Namespace, scalerNamespacedName.Name, resp.Status.Code, resp.Status.Message)
	}

	for _, pod := range resp.GetPods() {
		pNS := pod.GetObjectMeta().GetNamespace()
		pName := pod.GetObjectMeta().GetName()

		if pNS == "" && pName == "" {
			continue
		}

		podsInDatahub = append(podsInDatahub, &autoscalingv1alpha1.AlamedaPod{
			Namespace: pNS,
			Name:      pName,
		})
	}

	return podsInDatahub, nil
}

func (r *ReconcileAlamedaScaler) syncAlamedaRecommendation(alamedaScaler *autoscalingv1alpha1.AlamedaScaler, existingPodsMap map[autoscalingv1alpha1.NamespacedName]bool) error {

	currentPods := alamedaScaler.GetMonitoredPods()

	if err := r.createAssociateRecommendation(alamedaScaler, currentPods); err != nil {
		return errors.Wrapf(err, "sync AlamedaRecommendation failed: %s", err.Error())
	}

	if err := r.deleteAlamedaRecommendations(alamedaScaler, existingPodsMap); err != nil {
		return errors.Wrapf(err, "sync AlamedaRecommendation failed: %s", err.Error())
	}

	return nil
}

func (r *ReconcileAlamedaScaler) createAssociateRecommendation(alamedaScaler *autoscalingv1alpha1.AlamedaScaler, pods []*autoscalingv1alpha1.AlamedaPod) error {

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

		err := controllerutil.SetControllerReference(alamedaScaler, recommendation, r.scheme)
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

func (r *ReconcileAlamedaScaler) listAlamedaRecommendationsOwnedByAlamedaScaler(alamedaScaler *autoscalingv1alpha1.AlamedaScaler) ([]*autoscalingv1alpha1.AlamedaRecommendation, error) {

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

func (r *ReconcileAlamedaScaler) deleteAlamedaRecommendations(alamedaScaler *autoscalingv1alpha1.AlamedaScaler, existingPodsMap map[autoscalingv1alpha1.NamespacedName]bool) error {

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

func (r *ReconcileAlamedaScaler) getNeedDeletingAlamedaRecommendations(alamedaScaler *autoscalingv1alpha1.AlamedaScaler, existingPodsMap map[autoscalingv1alpha1.NamespacedName]bool) ([]*autoscalingv1alpha1.AlamedaRecommendation, error) {

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

func (r *ReconcileAlamedaScaler) getAlamedaScalerDatahubScalingType(alamedaScaler autoscalingv1alpha1.AlamedaScaler) datahub_resources.ScalingTool {
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
