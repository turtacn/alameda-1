/*
Copyright 2018 The Alameda Authors.

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

package alamedaresourceprediction

import (
	"context"
	"fmt"
	"reflect"

	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/pkg/apis/autoscaling/v1alpha1"
	logUtil "github.com/containers-ai/alameda/operator/pkg/utils/log"
	utilsresource "github.com/containers-ai/alameda/operator/pkg/utils/resources"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AlamedaResourcePrediction Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this autoscaling.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAlamedaResourcePrediction{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("alamedaresourceprediction-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to AlamedaResourcePrediction

	if err = c.Watch(&source.Kind{Type: &autoscalingv1alpha1.AlamedaResourcePrediction{}}, &handler.EnqueueRequestForObject{}); err != nil {
		logUtil.GetLogger().Error(err, fmt.Sprintf("Watch AlamedaResourcePrediction failed."))
		return err
	}
	if err = c.Watch(&source.Kind{Type: &autoscalingv1alpha1.AlamedaResource{}}, &handler.EnqueueRequestForObject{}); err != nil {
		logUtil.GetLogger().Error(err, fmt.Sprintf("Watch AlamedaResource controller for AlamedaResourcePrediction failed."))
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAlamedaResourcePrediction{}

// ReconcileAlamedaResourcePrediction reconciles a AlamedaResourcePrediction object
type ReconcileAlamedaResourcePrediction struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a AlamedaResourcePrediction object and makes changes based on the state read
// and what is in the AlamedaResourcePrediction.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=autoscaling.containers.ai,resources=alamedaresources,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling.containers.ai,resources=alamedaresourcepredictions,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileAlamedaResourcePrediction) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the AlamedaResourcePrediction instance
	instance := &autoscalingv1alpha1.AlamedaResourcePrediction{}

	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	copyInstance := &autoscalingv1alpha1.AlamedaResourcePrediction{}
	instance.DeepCopyInto(copyInstance)

	// Handle Alameda Deployment, maintain predict basic structure
	matchedDeploymentList := &appsv1.DeploymentList{}
	err = r.List(context.TODO(),
		client.InNamespace(instance.GetNamespace()).
			MatchingLabels(instance.Spec.Selector.MatchLabels),
		matchedDeploymentList)
	if err == nil {
		// 1. remove deleted deployment
		for deployUID, _ := range instance.Status.Prediction.Deployments {
			deleteDeploy := true
			for _, deployment := range matchedDeploymentList.Items {
				if string(deployment.GetUID()) == string(deployUID) {
					deleteDeploy = false
					break
				}
			}
			if deleteDeploy {
				delete(instance.Status.Prediction.Deployments, deployUID)
			}
		}
		for _, deployment := range matchedDeploymentList.Items {
			// 2. add new deployment
			if _, ok := instance.Status.Prediction.Deployments[autoscalingv1alpha1.DeploymentUID(deployment.GetUID())]; !ok {
				listPods := utilsresource.NewListPods(r)
				podList := listPods.ListPods(deployment.GetNamespace(), deployment.GetName(), "deployment")
				podsMap := map[autoscalingv1alpha1.PodUID]autoscalingv1alpha1.PredictPod{}
				for _, pod := range podList {
					containers := map[autoscalingv1alpha1.ContainerName]autoscalingv1alpha1.PredictContainer{}
					for _, container := range pod.Spec.Containers {
						containers[autoscalingv1alpha1.ContainerName(container.Name)] = autoscalingv1alpha1.PredictContainer{
							Name:            container.Name,
							RawPredict:      map[autoscalingv1alpha1.ResourceType]autoscalingv1alpha1.TimeSeriesData{},
							Recommendations: []autoscalingv1alpha1.Recommendation{},
						}
					}
					podsMap[autoscalingv1alpha1.PodUID(pod.GetUID())] = autoscalingv1alpha1.PredictPod{
						Name:       pod.GetName(),
						Containers: containers,
					}
				}

				instance.Status.Prediction.Deployments[autoscalingv1alpha1.DeploymentUID(deployment.GetUID())] = autoscalingv1alpha1.PredictDeployment{
					UID:       string(instance.GetUID()),
					Namespace: instance.GetNamespace(),
					Name:      instance.GetName(),
					Pods:      podsMap,
				}
			} else {
				// 3. update pods info of existing deployment
				listPods := utilsresource.NewListPods(r)
				podList := listPods.ListPods(deployment.GetNamespace(), deployment.GetName(), "deployment")
				// 3.1 remove deleted pods from existing deployment
				for podUID, _ := range instance.Status.Prediction.Deployments[autoscalingv1alpha1.DeploymentUID(deployment.GetUID())].Pods {
					deletePod := true
					for _, pod := range podList {
						if string(pod.GetUID()) == string(podUID) {
							deletePod = false
						}
					}
					if deletePod {
						delete(instance.Stsatus.Prediction.Deployments[autoscalingv1alpha1.DeploymentUID(deployment.GetUID())].Pods, podUID)
					}
				}

				for _, pod := range podList {
					// 3.2 add new pods from existing deployment
					if _, ok := instance.Status.Prediction.Deployments[autoscalingv1alpha1.DeploymentUID(deployment.GetUID())].Pods[autoscalingv1alpha1.PodUID(pod.GetUID())]; !ok {
						containers := map[autoscalingv1alpha1.ContainerName]autoscalingv1alpha1.PredictContainer{}
						for _, container := range pod.Spec.Containers {
							containers[autoscalingv1alpha1.ContainerName(container.Name)] = autoscalingv1alpha1.PredictContainer{
								Name:            container.Name,
								RawPredict:      map[autoscalingv1alpha1.ResourceType]autoscalingv1alpha1.TimeSeriesData{},
								Recommendations: []autoscalingv1alpha1.Recommendation{},
							}
						}
						instance.Status.Prediction.Deployments[autoscalingv1alpha1.DeploymentUID(deployment.GetUID())].Pods[autoscalingv1alpha1.PodUID(pod.GetUID())] = autoscalingv1alpha1.PredictPod{
							Name:       pod.GetName(),
							Containers: containers,
						}
					}
				}
			}
		}
		if !reflect.DeepEqual(instance.Status.Prediction, copyInstance.Status.Prediction) {
			logUtil.GetLogger().Info(fmt.Sprintf("Sync AlamedaResourcePrediction structure (%s/%s).", request.Namespace, request.Name))
			r.Update(context.TODO(), instance)
		}
	}

	return reconcile.Result{}, nil
}
