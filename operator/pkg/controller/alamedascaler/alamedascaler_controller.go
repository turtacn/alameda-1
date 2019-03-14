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
	"strings"
	"time"

	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/pkg/apis/autoscaling/v1alpha1"
	alamedascaler_reconciler "github.com/containers-ai/alameda/operator/pkg/reconciler/alamedascaler"
	"github.com/containers-ai/alameda/operator/pkg/utils"
	datahubutils "github.com/containers-ai/alameda/operator/pkg/utils/datahub"
	utilsresource "github.com/containers-ai/alameda/operator/pkg/utils/resources"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"github.com/pkg/errors"
	"google.golang.org/genproto/googleapis/rpc/code"
	"google.golang.org/grpc"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	scope = logUtil.RegisterScope("alamedascaler", "alamedascaler log", 0)
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
	return &ReconcileAlamedaScaler{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
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

	// Take care of AlamedaScaler
	if alamedaScaler, err := getResource.GetAlamedaScaler(request.Namespace, request.Name); err != nil && k8sErrors.IsNotFound(err) {
		scope.Infof("AlamedaScaler (%s/%s) is deleted, remove alameda pods from datahub.", request.Namespace, request.Name)
		err := deletePodsFromDatahub(&request.NamespacedName, make(map[autoscalingv1alpha1.NamespacedName]bool))
		if err != nil {
			scope.Errorf("Remove alameda pods of alamedascaler (%s/%s) from datahub failed. %s", request.Namespace, request.Name, err.Error())
		} else {
			scope.Infof("Remove alameda pods of alamedascaler (%s/%s) from datahub successed.", request.Namespace, request.Name)
		}
	} else if err == nil {
		// TODO: deployment already in the AlamedaScaler cannot join the other
		alamedaScalerNS := alamedaScaler.GetNamespace()
		alamedaScalerName := alamedaScaler.GetName()
		alamedascalerReconciler := alamedascaler_reconciler.NewReconciler(r, alamedaScaler)
		alamedaScaler, _ := alamedascalerReconciler.InitAlamedaController()
		alamedaScaler.ResetStatusAlamedaController()

		scope.Infof(fmt.Sprintf("AlamedaScaler (%s/%s) found, try to sync latest alamedacontrollers.", alamedaScalerNS, alamedaScalerName))
		if alamedaDeployments, err := listResources.ListDeploymentsByNamespaceLabels(request.Namespace, alamedaScaler.Spec.Selector.MatchLabels); err == nil {
			for _, alamedaDeployment := range alamedaDeployments {
				alamedaScaler = alamedascalerReconciler.UpdateStatusByDeployment(&alamedaDeployment)
			}
		} else {
			scope.Error(err.Error())
		}

		if alamedaDeploymentConfigs, err := listResources.ListDeploymentConfigsByNamespaceLabels(request.Namespace, alamedaScaler.Spec.Selector.MatchLabels); err == nil {
			for _, alamedaDeploymentConfig := range alamedaDeploymentConfigs {
				alamedaScaler = alamedascalerReconciler.UpdateStatusByDeploymentConfig(&alamedaDeploymentConfig)
			}
		} else {
			scope.Error(err.Error())
		}

		if err := updateResource.UpdateAlamedaScaler(alamedaScaler); err != nil {
			scope.Errorf("update AlamedaScaler %s/%s failed: %s", alamedaScalerNS, alamedaScalerName, err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
		}

		// after updating AlamedaPod in AlamedaScaler, start create AlamedaRecommendation if necessary and register alameda pod to datahub
		if err := r.syncAlamedaScalerWithDepResources(alamedaScaler); err != nil {
			scope.Error(err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: 1 * time.Second}, nil
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

	numOfGoroutine := 2
	done := make(chan bool)
	errChan := make(chan error)
	go r.syncDatahubResource(done, errChan, alamedaScaler, existingPodsMap)
	go r.syncAlamedaRecommendation(done, errChan, alamedaScaler, existingPodsMap)

	for i := 0; i < numOfGoroutine; i++ {
		select {
		case _ = <-done:
			continue
		case err := <-errChan:
			if err != nil {
				return errors.Wrapf(err, "sync AlamedaScaler %s/%s with dependent resources failed: %s", alamedaScaler.Namespace, alamedaScaler.Name, err.Error())
			}
		}
	}

	return nil
}

func (r *ReconcileAlamedaScaler) syncDatahubResource(done chan bool, errChan chan error, alamedaScaler *autoscalingv1alpha1.AlamedaScaler, existingPodsMap map[autoscalingv1alpha1.NamespacedName]bool) error {

	currentPods := alamedaScaler.GetMonitoredPods()

	if err := r.createPodsToDatahub(alamedaScaler, currentPods); err != nil {
		errChan <- errors.Wrapf(err, "sync Datahub resource failed: %s", err.Error())
	}

	if err := deletePodsFromDatahub(&types.NamespacedName{
		Namespace: alamedaScaler.GetNamespace(),
		Name:      alamedaScaler.GetName(),
	}, existingPodsMap); err != nil {
		errChan <- errors.Wrapf(err, "sync Datahub resource failed: %s", err.Error())
	}

	done <- true
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

	policy := datahub_v1alpha1.RecommendationPolicy_STABLE
	if strings.ToLower(string(scaler.Spec.Policy)) == strings.ToLower(string(autoscalingv1alpha1.RecommendationPolicyCOMPACT)) {
		policy = datahub_v1alpha1.RecommendationPolicy_COMPACT
	} else if strings.ToLower(string(scaler.Spec.Policy)) == strings.ToLower(string(autoscalingv1alpha1.RecommendationPolicySTABLE)) {
		policy = datahub_v1alpha1.RecommendationPolicy_STABLE
	}

	podsNeedCreating := []*datahub_v1alpha1.Pod{}
	for _, pod := range pods {
		containers := []*datahub_v1alpha1.Container{}
		startTime := &timestamp.Timestamp{}
		for _, container := range pod.Containers {
			containers = append(containers, &datahub_v1alpha1.Container{
				Name: container.Name,
			})
		}
		nodeName := ""
		resourceLink := ""
		if corePod, err := getResource.GetPod(pod.Namespace, pod.Name); err == nil {
			nodeName = corePod.Spec.NodeName
			startTime = &timestamp.Timestamp{
				Seconds: corePod.ObjectMeta.GetCreationTimestamp().Unix(),
			}
			resourceLink = utilsresource.GetResourceLinkForPod(r.Client, corePod)
			scope.Infof(fmt.Sprintf("resource link for pod (%s/%s) is %s", corePod.GetNamespace(), corePod.GetName(), resourceLink))
		} else {
			scope.Errorf("build Datahub pod to create failed, skip this pod: get pod %s/%s from k8s failed: %s", pod.Namespace, pod.Name, err.Error())
			continue
		}

		topCtrl, err := utils.ParseResourceLinkForTopController(resourceLink)
		if err != nil {
			scope.Error(err.Error())
		}

		podsNeedCreating = append(podsNeedCreating, &datahub_v1alpha1.Pod{
			AlamedaScaler: &datahub_v1alpha1.NamespacedName{
				Namespace: scaler.Namespace,
				Name:      scaler.Name,
			},
			NamespacedName: &datahub_v1alpha1.NamespacedName{
				Namespace: pod.Namespace,
				Name:      pod.Name,
			},
			Policy:        datahub_v1alpha1.RecommendationPolicy(policy),
			Containers:    containers,
			NodeName:      nodeName,
			ResourceLink:  resourceLink,
			StartTime:     startTime,
			TopController: topCtrl,
		})
	}

	req := datahub_v1alpha1.CreatePodsRequest{
		Pods: podsNeedCreating,
	}
	resp, err := datahubServiceClnt.CreatePods(context.Background(), &req)
	if err != nil {
		return errors.Errorf("add alameda pods for AlamedaScaler (%s/%s) failed: %s", scaler.GetNamespace(), scaler.GetName(), err.Error())
	} else if resp.Code != int32(code.Code_OK) {
		return errors.Errorf("add alameda pods for AlamedaScaler (%s/%s) failed: receive response: code: %d, message: %s", scaler.GetNamespace(), scaler.GetName(), resp.Code, resp.Message)
	}
	scope.Infof(fmt.Sprintf("add alameda pods for AlamedaScaler (%s/%s) successfully", scaler.GetNamespace(), scaler.GetName()))

	return nil
}

func deletePodsFromDatahub(scalerNamespacedName *types.NamespacedName, existingPodsMap map[autoscalingv1alpha1.NamespacedName]bool) error {

	pods, err := getPodsNeedDeleting(scalerNamespacedName, existingPodsMap)
	if err != nil {
		return errors.Wrapf(err, "delete pods from datahub failed: %s", err.Error())
	}

	conn, err := grpc.Dial(datahubutils.GetDatahubAddress(), grpc.WithInsecure())
	if err != nil {
		return errors.Errorf("delete pods from datahub failed: %s", err.Error())
	}

	defer conn.Close()
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(conn)

	podsNeedDeleting := []*datahub_v1alpha1.Pod{}
	for _, pod := range pods {
		podsNeedDeleting = append(podsNeedDeleting, &datahub_v1alpha1.Pod{
			NamespacedName: &datahub_v1alpha1.NamespacedName{
				Namespace: pod.Namespace,
				Name:      pod.Name,
			},
			AlamedaScaler: &datahub_v1alpha1.NamespacedName{
				Namespace: scalerNamespacedName.Namespace,
				Name:      scalerNamespacedName.Name,
			},
		})
	}

	req := datahub_v1alpha1.DeletePodsRequest{
		Pods: podsNeedDeleting,
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

func getPodsNeedDeleting(scalerNamespacedName *types.NamespacedName, existingPodsMap map[autoscalingv1alpha1.NamespacedName]bool) ([]*autoscalingv1alpha1.AlamedaPod, error) {

	copyScaler := *scalerNamespacedName

	needDeletingPods := make([]*autoscalingv1alpha1.AlamedaPod, 0)
	podsInDatahub, err := getPodsObservedByAlamedaScalerFromDatahub(&copyScaler)
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

func getPodsObservedByAlamedaScalerFromDatahub(scalerNamespacedName *types.NamespacedName) ([]*autoscalingv1alpha1.AlamedaPod, error) {

	podsInDatahub := make([]*autoscalingv1alpha1.AlamedaPod, 0)

	conn, err := grpc.Dial(datahubutils.GetDatahubAddress(), grpc.WithInsecure())
	if err != nil {
		return podsInDatahub, errors.Errorf("get pods from datahub failed: %s", err.Error())
	}

	defer conn.Close()
	datahubServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(conn)

	req := datahub_v1alpha1.ListAlamedaPodsRequest{
		NamespacedName: &datahub_v1alpha1.NamespacedName{
			Namespace: scalerNamespacedName.Namespace,
			Name:      scalerNamespacedName.Name,
		},
		Kind: datahub_v1alpha1.Kind_ALAMEDASCALER,
	}
	resp, err := datahubServiceClnt.ListAlamedaPods(context.Background(), &req)
	if err != nil {
		return podsInDatahub, errors.Errorf("get alameda pods for AlamedaScaler (%s/%s) failed: %s", scalerNamespacedName.Namespace, scalerNamespacedName.Name, err.Error())
	} else if resp.Status == nil {
		return podsInDatahub, errors.Errorf("get alameda pods for AlamedaScaler (%s/%s) failed: receive null status", scalerNamespacedName.Namespace, scalerNamespacedName.Name)
	} else if resp.Status.Code != int32(code.Code_OK) {
		return podsInDatahub, errors.Errorf("get alameda pods for AlamedaScaler (%s/%s) failed: receive response: code: %d, message: %s", scalerNamespacedName.Namespace, scalerNamespacedName.Name, resp.Status.Code, resp.Status.Message)
	}

	for _, pod := range resp.GetPods() {

		namespacedName := pod.GetNamespacedName()
		if namespacedName == nil {
			continue
		}

		podsInDatahub = append(podsInDatahub, &autoscalingv1alpha1.AlamedaPod{
			Namespace: namespacedName.Namespace,
			Name:      namespacedName.Name,
		})
	}

	return podsInDatahub, nil
}

func (r *ReconcileAlamedaScaler) syncAlamedaRecommendation(done chan bool, errChan chan error, alamedaScaler *autoscalingv1alpha1.AlamedaScaler, existingPodsMap map[autoscalingv1alpha1.NamespacedName]bool) error {

	currentPods := alamedaScaler.GetMonitoredPods()

	if err := r.createAssociateRecommendation(alamedaScaler, currentPods); err != nil {
		return errors.Wrapf(err, "sync AlamedaRecommendation failed: %s", err.Error())
	}

	if err := r.deleteAlamedaRecommendations(alamedaScaler, existingPodsMap); err != nil {
		return errors.Wrapf(err, "sync AlamedaRecommendation failed: %s", err.Error())
	}

	done <- true
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
