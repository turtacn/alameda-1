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

package deploymentconfig

import (
	"time"

	alamedascaler_reconciler "github.com/containers-ai/alameda/operator/pkg/reconciler/alamedascaler"
	utilsresource "github.com/containers-ai/alameda/operator/pkg/utils/resources"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	appsapi_v1 "github.com/openshift/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	scope             = logUtil.RegisterScope("deploymentconfig_controller", "deploymentconfig controller log", 0)
	cachedFirstSynced = false
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Deployment Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileDeploymentConfig{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("deploymentconfig-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to DeploymentConfig
	err = c.Watch(&source.Kind{Type: &appsapi_v1.DeploymentConfig{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileDeploymentConfig{}

// ReconcileDeploymentConfig reconciles a DeploymentConfig object
type ReconcileDeploymentConfig struct {
	client.Client
	scheme *runtime.Scheme
}

func (r *ReconcileDeploymentConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	if !cachedFirstSynced {
		time.Sleep(5 * time.Second)
	}
	cachedFirstSynced = true

	getResource := utilsresource.NewGetResource(r)
	listResources := utilsresource.NewListResources(r)
	updateResource := utilsresource.NewUpdateResource(r)

	scalers, _ := listResources.ListAllAlamedaScaler()

	if deploymentConfig, err := getResource.GetDeploymentConfig(request.Namespace, request.Name); err != nil && errors.IsNotFound(err) {
	} else if err == nil {
		for _, alamedascaler := range scalers {
			alamedascalerReconciler := alamedascaler_reconciler.NewReconciler(r, &alamedascaler)
			matchedLblDeploymentConfigs, _ := listResources.ListDeploymentConfigsByLabels(alamedascaler.Spec.Selector.MatchLabels)
			for _, matchedLblDeploymentConfig := range matchedLblDeploymentConfigs {
				// deploymentconfig can only join one AlamedaScaler
				if matchedLblDeploymentConfig.GetUID() == deploymentConfig.GetUID() {
					alamedascaler = *alamedascalerReconciler.UpdateStatusByDeploymentConfig(deploymentConfig)
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
