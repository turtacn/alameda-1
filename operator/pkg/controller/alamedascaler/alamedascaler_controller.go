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

	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/pkg/apis/autoscaling/v1alpha1"
	alamedascaler_reconciler "github.com/containers-ai/alameda/operator/pkg/reconciler/alamedascaler"
	datahubutils "github.com/containers-ai/alameda/operator/pkg/utils/datahub"
	utilsresource "github.com/containers-ai/alameda/operator/pkg/utils/resources"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	timestamp "github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/grpc"
	appsv1 "k8s.io/api/apps/v1"

	logUtil "github.com/containers-ai/alameda/operator/pkg/utils/log"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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

// AlamedaScaler is alameda scaler
type AlamedaScaler string

//
const (
	AlamedaDeployment AlamedaScaler = "Deployment"
	UpdateRetry                     = 3
)

// AlamedaK8sController is key of AlamedaScaler annotation
const AlamedaK8sController = "annotation-k8s-controller"

// JSONIndent is ident of formatted json string
const JSONIndent = "  "

// Container struct
type Container struct {
	Name string
}

// Pod struct
type Pod struct {
	UID        string
	Namespace  string
	Name       string
	Containers []Container
}

// Deployment struct
type Deployment struct {
	UID       string
	Namespace string
	Name      string
	PodMap    map[string]Pod
}

// K8SControllerAnnotation struct
type K8SControllerAnnotation struct {
	DeploymentMap map[string]Deployment
}

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
	}

	if err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForObject{}); err != nil {
		scope.Error(err.Error())
	}

	if err != nil {
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
	if alamedaScaler, err := getResource.GetAlamedaScaler(request.Namespace, request.Name); err != nil && errors.IsNotFound(err) {
	} else if err == nil {
		// TODO: deployment already in the AlamedaScaler cannot join the other
		alamedaScalerNS := alamedaScaler.GetNamespace()
		alamedaScalerName := alamedaScaler.GetName()
		alamedascalerReconciler := alamedascaler_reconciler.NewReconciler(r, alamedaScaler)
		if alamedaScaler, needUpdated := alamedascalerReconciler.InitAlamedaController(); needUpdated {
			updateResource.UpdateAlamedaScaler(alamedaScaler)
		}

		scope.Infof(fmt.Sprintf("AlamedaScaler (%s/%s) found, try to sync latest alamedacontrollers.", alamedaScalerNS, alamedaScalerName))
		if alamedaDeployments, err := listResources.ListDeploymentsByLabels(alamedaScaler.Spec.Selector.MatchLabels); err == nil {
			for _, alamedaDeployment := range alamedaDeployments {
				alamedaScaler = alamedascalerReconciler.UpdateStatusByDeployment(&alamedaDeployment)
			}
			updateResource.UpdateAlamedaScaler(alamedaScaler)
		}

		// after updating AlamedaPod in AlamedaScaler, start create AlamedaRecommendation if necessary and register alameda pod to datahub
		if conn, err := grpc.Dial(datahubutils.GetDatahubAddress(), grpc.WithInsecure()); err == nil {
			defer conn.Close()
			aiServiceClnt := datahub_v1alpha1.NewDatahubServiceClient(conn)
			pods := []*datahub_v1alpha1.Pod{}
			policy := datahub_v1alpha1.RecommendationPolicy_STABLE
			if strings.ToLower(string(alamedaScaler.Spec.Policy)) == strings.ToLower(string(autoscalingv1alpha1.RecommendationPolicyCOMPACT)) {
				policy = datahub_v1alpha1.RecommendationPolicy_COMPACT
			} else if strings.ToLower(string(alamedaScaler.Spec.Policy)) == strings.ToLower(string(autoscalingv1alpha1.RecommendationPolicySTABLE)) {
				policy = datahub_v1alpha1.RecommendationPolicy_STABLE
			}
			for _, scalerDeployment := range alamedaScaler.Status.AlamedaController.Deployments {
				for _, pod := range scalerDeployment.Pods {
					containers := []*datahub_v1alpha1.Container{}
					startTime := &timestamp.Timestamp{}
					for _, container := range pod.Containers {
						containers = append(containers, &datahub_v1alpha1.Container{
							Name: container.Name,
						})
					}
					nodeName := ""
					if pod, err := getResource.GetPod(scalerDeployment.Namespace, pod.Name); err == nil {
						nodeName = pod.Spec.NodeName
						startTime = &timestamp.Timestamp{
							Seconds: pod.ObjectMeta.GetCreationTimestamp().Unix(),
						}
					} else {
						scope.Error(err.Error())
					}

					pods = append(pods, &datahub_v1alpha1.Pod{
						IsAlameda: true,
						AlamedaScaler: &datahub_v1alpha1.NamespacedName{
							Namespace: alamedaScalerNS,
							Name:      alamedaScalerName,
						},
						NamespacedName: &datahub_v1alpha1.NamespacedName{
							Namespace: scalerDeployment.Namespace,
							Name:      pod.Name,
						},
						Policy:     datahub_v1alpha1.RecommendationPolicy(policy),
						Containers: containers,
						NodeName:   nodeName,
						// TODO
						ResourceLink: "",
						StartTime:    startTime,
					})
					// try to create the recommendation by pod
					recommendationNS := scalerDeployment.Namespace
					recommendationName := pod.Name

					recommendation := &autoscalingv1alpha1.AlamedaRecommendation{
						ObjectMeta: metav1.ObjectMeta{
							Name:      recommendationName,
							Namespace: recommendationNS,
							Labels: map[string]string{
								"alamedascaler": fmt.Sprintf("%s.%s", alamedaScaler.GetName(), alamedaScaler.GetNamespace()),
							},
						},
						Spec: autoscalingv1alpha1.AlamedaRecommendationSpec{
							Containers: pod.Containers,
						},
					}

					if err := controllerutil.SetControllerReference(alamedaScaler, recommendation, r.scheme); err == nil {
						_, err := getResource.GetAlamedaRecommendation(recommendationNS, recommendationName)
						if err != nil && errors.IsNotFound(err) {
							err = r.Create(context.TODO(), recommendation)
							if err != nil {
								scope.Error(err.Error())
							}
						}
					}
				}
			}
			req := datahub_v1alpha1.CreatePodsRequest{
				Pods: pods,
			}
			_, err := aiServiceClnt.CreatePods(context.Background(), &req)
			if err != nil {
				scope.Error(err.Error())
			} else {
				scope.Infof(fmt.Sprintf("Add/Update alameda pods for AlamedaScaler (%s/%s) successfully", alamedaScaler.GetNamespace(), alamedaScaler.GetName()))
			}
		} else {
			scope.Error(err.Error())
		}
	} else {
		scope.Error(err.Error())
	}

	// Take care of Deployment
	allAlamedaScalers, _ := listResources.ListAllAlamedaScaler()
	for _, alamedascaler := range allAlamedaScalers {
		alamedascalerReconciler := alamedascaler_reconciler.NewReconciler(r, &alamedascaler)
		if alamedascalerReconciler.HasAlamedaDeployment(request.Namespace, request.Name) {
			updateResource.UpdateAlamedaScaler(&alamedascaler)
		}
	}
	if deployment, err := getResource.GetDeployment(request.Namespace, request.Name); err != nil && errors.IsNotFound(err) {
	} else if err == nil {
		for _, alamedascaler := range allAlamedaScalers {
			alamedascalerReconciler := alamedascaler_reconciler.NewReconciler(r, &alamedascaler)
			matchedLblDeployments, _ := listResources.ListDeploymentsByLabels(alamedascaler.Spec.Selector.MatchLabels)
			for _, matchedLblDeployment := range matchedLblDeployments {
				// deployment can only join one AlamedaScaler
				if matchedLblDeployment.GetUID() == deployment.GetUID() {
					alamedascaler = *alamedascalerReconciler.UpdateStatusByDeployment(deployment)
					updateResource.UpdateAlamedaScaler(&alamedascaler)
					return reconcile.Result{}, nil
				}
			}
		}
	} else {
		scope.Error(err.Error())
	}

	return reconcile.Result{}, nil
}
