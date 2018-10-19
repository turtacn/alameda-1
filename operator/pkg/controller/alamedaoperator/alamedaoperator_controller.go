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

package alamedaoperator

import (
	"context"
	"fmt"
	"log"
	"reflect"
	"strings"

	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/pkg/apis/autoscaling/v1alpha1"
	logUtil "github.com/containers-ai/alameda/operator/pkg/utils/log"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

const (
	alamedaTag = "alameda"
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new AlamedaOperator Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
// USER ACTION REQUIRED: update cmd/manager/main.go to call this autoscaling.Add(mgr) to install this Controller
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileAlamedaOperator{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("alamedaoperator-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to AlamedaOperator
	err = c.Watch(&source.Kind{Type: &autoscalingv1alpha1.AlamedaOperator{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by AlamedaOperator - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &autoscalingv1alpha1.AlamedaOperator{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileAlamedaOperator{}

// ReconcileAlamedaOperator reconciles a AlamedaOperator object
type ReconcileAlamedaOperator struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a AlamedaOperator object and makes changes based on the state read
// and what is in the AlamedaOperator.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=autoscaling.containers.ai,resources=alamedaoperators,verbs=get;list;watch;create;update;patch;delete
func (r *ReconcileAlamedaOperator) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the AlamedaOperator instance
	instance := &autoscalingv1alpha1.AlamedaOperator{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	logUtil.GetLogger().Info(fmt.Sprintf("Get Alameda Resource %s/%s.", instance.Namespace, instance.Name))
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	policy := instance.Spec.Policy
	// TODO(user): Change this to be the object type created by your controller
	// Define the desired Deployment object
	deploy := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      instance.Name + "-" + alamedaTag,
			Namespace: instance.Namespace,
		},
		Spec: instance.Spec.DeploymentSpec,
	}
	if err := controllerutil.SetControllerReference(instance, deploy, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// TODO(user): Change this for the object type created by your controller
	// Check if the Deployment already exists
	found := &appsv1.Deployment{}
	err = r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		logUtil.GetLogger().Info(fmt.Sprintf("Creating Alameda Deployment with policy %s %s/%s.", policy, deploy.Namespace, deploy.Name))
		err = r.Create(context.TODO(), deploy)
		if err != nil {
			logUtil.GetLogger().Error(err, err.Error())
			return reconcile.Result{}, err
		}
	} else if err != nil {
		return reconcile.Result{}, err
	}

	r.getPodsFromDeployment(found)
	// TODO(user): Change this for the object type created by your controller
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(deploy.Spec, found.Spec) {
		found.Spec = deploy.Spec
		log.Printf("Updating Deployment %s/%s\n", deploy.Namespace, deploy.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return reconcile.Result{}, err
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileAlamedaOperator) getPodsFromDeployment(deployment *appsv1.Deployment) {
	pods := &corev1.PodList{}
	name := deployment.Name
	ns := deployment.GetNamespace()
	labels := deployment.GetLabels()

	err := r.Client.List(context.TODO(),
		client.InNamespace(ns).
			MatchingLabels(labels),
		pods)
	podUIDs := []string{}
	if err != nil {
		logUtil.GetLogger().Info(fmt.Sprintf("List pods of alameda deployment %s/%s failed.", ns, name))
	} else {
		var deploymentName string
		for _, pod := range pods.Items {
			for _, ownerReference := range pod.ObjectMeta.GetOwnerReferences() {

				if ownerReference.Kind == "ReplicaSet" {
					replicaSetName := ownerReference.Name
					deploymentName = replicaSetName[0:strings.LastIndex(replicaSetName, "-")]
				}
				break
			}
			if deploymentName == name {
				podUIDs = append(podUIDs, string(pod.GetUID()))
			}
		}
	}
	logUtil.GetLogger().Info(fmt.Sprintf("%d pods founded in alameda deployment %s/%s.", len(podUIDs), ns, name))
}
