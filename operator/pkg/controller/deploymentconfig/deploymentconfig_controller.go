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

	datahub_client_controller "github.com/containers-ai/alameda/operator/datahub/client/controller"
	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/pkg/apis/autoscaling/v1alpha1"
	controllerutil "github.com/containers-ai/alameda/operator/pkg/controller/util"
	datahubutils "github.com/containers-ai/alameda/operator/pkg/utils/datahub"
	utilsresource "github.com/containers-ai/alameda/operator/pkg/utils/resources"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	appsapi_v1 "github.com/openshift/api/apps/v1"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
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
	requeueDuration   = 1 * time.Second
	grpcDefaultRetry  = uint(3)
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
	conn, _ := grpc.Dial(datahubutils.GetDatahubAddress(), grpc.WithInsecure(),
		grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(grpc_retry.WithMax(grpcDefaultRetry))))
	datahubControllerRepo := datahub_client_controller.NewControllerRepository(conn)
	return &ReconcileDeploymentConfig{
		Client:                mgr.GetClient(),
		scheme:                mgr.GetScheme(),
		datahubControllerRepo: datahubControllerRepo,
	}
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

	datahubControllerRepo *datahub_client_controller.ControllerRepository
}

func (r *ReconcileDeploymentConfig) Reconcile(request reconcile.Request) (reconcile.Result, error) {

	if !cachedFirstSynced {
		time.Sleep(5 * time.Second)
	}
	cachedFirstSynced = true

	getResource := utilsresource.NewGetResource(r)
	updateResource := utilsresource.NewUpdateResource(r)

	deploymentConfig := appsapi_v1.DeploymentConfig{}
	err := r.Get(context.Background(), request.NamespacedName, &deploymentConfig)
	if err != nil && k8sErrors.IsNotFound(err) {
		// If deploymentConfig is deleted, it cannnot find the monitoring AlamedaScaler by calling method GetObservingAlamedaScalerOfController
		// in type GetResource.
		alamedaScaler, err := r.getMonitoringAlamedaScaler(request.Namespace, request.Name)
		if err != nil {
			scope.Errorf("Get observing AlamedaScaler of DeploymentConfig failed: %s", err.Error())
			return reconcile.Result{}, nil
		} else if alamedaScaler == nil {
			scope.Warnf("Get observing AlamedaScaler of DeploymentConfig %s/%s not found", request.Namespace, request.Name)
			return reconcile.Result{}, nil
		}

		alamedaScaler.SetCustomResourceVersion(alamedaScaler.GenCustomResourceVersion())
		err = updateResource.UpdateAlamedaScaler(alamedaScaler)
		if err != nil {
			scope.Errorf("Update AlamedaScaler falied: %s", err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: requeueDuration}, nil
		}

		// delete controller to datahub
		err = r.datahubControllerRepo.DeleteControllers([]*datahub_resources.Controller{
			&datahub_resources.Controller{
				ObjectMeta: &datahub_resources.ObjectMeta{
					Name:      request.NamespacedName.Name,
					Namespace: request.NamespacedName.Namespace,
				},
				Kind: datahub_resources.Kind_STATEFULSET,
			},
		}, nil)
		if err != nil {
			scope.Errorf("Delete controller %s/%s from datahub failed: %s",
				request.NamespacedName.Namespace, request.NamespacedName.Name, err.Error())
		}
	} else if err != nil {
		scope.Errorf("Get DeploymentConfig %s/%s failed: %s", request.Namespace, request.Name, err.Error())
		return reconcile.Result{}, nil
	} else {
		alamedaScaler, err := getResource.GetObservingAlamedaScalerOfController(autoscalingv1alpha1.DeploymentConfigController, request.Namespace, request.Name)
		if err != nil && !k8sErrors.IsNotFound(err) {
			scope.Errorf("Get observing AlamedaScaler of DeploymentConfig failed: %s", err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: requeueDuration}, nil
		} else if alamedaScaler == nil {
			scope.Warnf("Get observing AlamedaScaler of DeploymentConfig %s/%s not found", request.Namespace, request.Name)
		}

		var currentMonitorAlamedaScalerName = ""
		if alamedaScaler != nil {
			if err := controllerutil.TriggerAlamedaScaler(updateResource, alamedaScaler); err != nil {
				scope.Errorf("Trigger current monitoring AlamedaScaler to update falied: %s", err.Error())
				return reconcile.Result{Requeue: true, RequeueAfter: requeueDuration}, nil
			}
			currentMonitorAlamedaScalerName = alamedaScaler.Name
		}

		lastMonitorAlamedaScalerName := controllerutil.GetLastMonitorAlamedaScaler(&deploymentConfig)
		// Do not trigger the update process twice if last and current AlamedaScaler are the same
		if lastMonitorAlamedaScalerName != "" && currentMonitorAlamedaScalerName != lastMonitorAlamedaScalerName {
			lastMonitorAlamedaScaler, err := getResource.GetAlamedaScaler(request.Namespace, lastMonitorAlamedaScalerName)
			if err != nil && !k8sErrors.IsNotFound(err) {
				scope.Errorf("Get last monitoring AlamedaScaler falied: %s", err.Error())
				return reconcile.Result{Requeue: true, RequeueAfter: requeueDuration}, nil
			}
			if lastMonitorAlamedaScaler != nil {
				err := controllerutil.TriggerAlamedaScaler(updateResource, lastMonitorAlamedaScaler)
				if err != nil {
					scope.Errorf("Trigger last monitoring AlamedaScaler to update falied: %s", err.Error())
					return reconcile.Result{Requeue: true, RequeueAfter: requeueDuration}, nil
				}
			}
		}

		controllerutil.SetLastMonitorAlamedaScaler(&deploymentConfig, currentMonitorAlamedaScalerName)
		err = updateResource.UpdateResource(&deploymentConfig)
		if err != nil {
			scope.Errorf("Update DeploymentConfig falied: %s", err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: requeueDuration}, nil
		}

		// add controller to datahub
		err = r.datahubControllerRepo.CreateControllers([]*datahub_resources.Controller{
			&datahub_resources.Controller{
				ObjectMeta: &datahub_resources.ObjectMeta{
					Name:      request.NamespacedName.Name,
					Namespace: request.NamespacedName.Namespace,
				},
				Kind: datahub_resources.Kind_STATEFULSET,
			},
		})
		if err != nil {
			scope.Errorf("Create controller %s/%s from datahub failed: %s",
				request.NamespacedName.Namespace, request.NamespacedName.Name, err.Error())
		}
	}
	return reconcile.Result{}, nil
}

func (r *ReconcileDeploymentConfig) getMonitoringAlamedaScaler(namespace, name string) (*autoscalingv1alpha1.AlamedaScaler, error) {

	listResource := utilsresource.NewListResources(r.Client)
	alamedaScalers, err := listResource.ListNamespaceAlamedaScaler(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "list AlamedaScalers failed")
	}

	for _, alamedaScaler := range alamedaScalers {
		for _, deployment := range alamedaScaler.Status.AlamedaController.DeploymentConfigs {
			if deployment.Namespace == namespace && deployment.Name == name {
				return &alamedaScaler, nil
			}
		}
	}

	return nil, nil
}
