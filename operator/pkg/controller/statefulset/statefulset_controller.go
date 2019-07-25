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

package statefulset

import (
	"context"
	"time"

	"github.com/pkg/errors"

	autoscaling_v1alpha1 "github.com/containers-ai/alameda/operator/pkg/apis/autoscaling/v1alpha1"
	utilsresource "github.com/containers-ai/alameda/operator/pkg/utils/resources"
	"github.com/containers-ai/alameda/pkg/utils/log"

	appsv1 "k8s.io/api/apps/v1"
	k8s_errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	scope           = log.RegisterScope("statefulset_controller", "deployment controller log", 0)
	requeueDuration = 1 * time.Second
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new StatefulSet Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileStatefulSet{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("statefulset-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to StatefulSet
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileStatefulSet{}

// ReconcileStatefulSet reconciles a StatefulSet object
type ReconcileStatefulSet struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a StatefulSet object and makes changes based on the state read
// and what is in the StatefulSet.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a StatefulSet as an example
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets/status,verbs=get;update;patch
func (r *ReconcileStatefulSet) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	getResource := utilsresource.NewGetResource(r)
	updateResource := utilsresource.NewUpdateResource(r)

	statefulSet := appsv1.StatefulSet{}
	err := r.Get(context.Background(), request.NamespacedName, &statefulSet)
	if err != nil && k8s_errors.IsNotFound(err) {
		// If statefulSet is deleted, it cannnot find the monitoring AlamedaScaler by calling method GetObservingAlamedaScalerOfController
		// in type GetResource.
		alamedaScaler, err := r.getMonitoringAlamedaScaler(request.Namespace, request.Name)
		if err != nil {
			scope.Errorf("Get observing AlamedaScaler of StatefulSet failed: %s", err.Error())
			return reconcile.Result{}, nil
		} else if alamedaScaler == nil {
			scope.Warnf("Observing AlamedaScaler of StatefulSet %s/%s not found", request.Namespace, request.Name)
			return reconcile.Result{}, nil
		}

		alamedaScaler.SetCustomResourceVersion(alamedaScaler.GenCustomResourceVersion())
		err = updateResource.UpdateAlamedaScaler(alamedaScaler)
		if err != nil {
			scope.Errorf("Update AlamedaScaler falied: %s", err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: requeueDuration}, nil
		}
	} else if err != nil {
		scope.Errorf("Get StatefulSet %s/%s failed: %s", request.Namespace, request.Name, err.Error())
		return reconcile.Result{}, nil
	} else {
		alamedaScaler, err := getResource.GetObservingAlamedaScalerOfController(autoscaling_v1alpha1.StatefulSetController, request.Namespace, request.Name)
		if err != nil {
			scope.Errorf("Get observing AlamedaScaler of deployment failed: %s", err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: requeueDuration}, nil
		} else if alamedaScaler == nil {
			scope.Warnf("Get observing AlamedaScaler of deployment %s/%s not found", request.Namespace, request.Name)
			return reconcile.Result{}, nil
		}

		alamedaScaler.SetCustomResourceVersion(alamedaScaler.GenCustomResourceVersion())
		err = updateResource.UpdateAlamedaScaler(alamedaScaler)
		if err != nil {
			scope.Errorf("Update AlamedaScaler falied: %s", err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: requeueDuration}, nil
		}
	}

	return reconcile.Result{}, nil

}

func (r *ReconcileStatefulSet) getMonitoringAlamedaScaler(namespace, name string) (*autoscaling_v1alpha1.AlamedaScaler, error) {

	listResource := utilsresource.NewListResources(r.Client)
	alamedaScalers, err := listResource.ListNamespaceAlamedaScaler(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "list AlamedaScalers failed")
	}

	for _, alamedaScaler := range alamedaScalers {
		for _, statefulSet := range alamedaScaler.Status.AlamedaController.StatefulSets {
			if statefulSet.Namespace == namespace && statefulSet.Name == name {
				return &alamedaScaler, nil
			}
		}
	}

	return nil, nil
}
