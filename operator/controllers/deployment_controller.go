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
	"time"

	autoscalingv1alpha1 "github.com/containers-ai/alameda/operator/api/v1alpha1"
	controllerutil "github.com/containers-ai/alameda/operator/controllers/util"
	datahub_client_controller "github.com/containers-ai/alameda/operator/datahub/client/controller"
	utilsresource "github.com/containers-ai/alameda/operator/pkg/utils/resources"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	appsv1 "k8s.io/api/apps/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// DeploymentReconciler reconciles a Deployment object
type DeploymentReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	DatahubControllerRepo *datahub_client_controller.ControllerRepository

	ClusterUID string
}

func (r *DeploymentReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	requeueDuration := 1 * time.Second
	getResource := utilsresource.NewGetResource(r)
	updateResource := utilsresource.NewUpdateResource(r)

	deployment := appsv1.Deployment{}
	err := r.Get(context.Background(), req.NamespacedName, &deployment)
	if err != nil && k8sErrors.IsNotFound(err) {
		// If deployment is deleted, it cannnot find the monitoring AlamedaScaler by calling method GetObservingAlamedaScalerOfController
		// in type GetResource.
		alamedaScaler, err := r.getMonitoringAlamedaScaler(req.Namespace, req.Name)
		if err != nil {
			scope.Errorf("Get observing AlamedaScaler of Deployment failed: %s", err.Error())
			return ctrl.Result{}, nil
		} else if alamedaScaler == nil {
			scope.Warnf("Get observing AlamedaScaler of Deployment %s/%s not found", req.Namespace, req.Name)
			return ctrl.Result{}, nil
		}

		alamedaScaler.SetCustomResourceVersion(alamedaScaler.GenCustomResourceVersion())
		err = updateResource.UpdateAlamedaScaler(alamedaScaler)
		if err != nil {
			scope.Errorf("Update AlamedaScaler falied: %s", err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: requeueDuration}, nil
		}

		// delete controller to datahub
		err = r.DatahubControllerRepo.DeleteControllers(context.TODO(), []*datahub_resources.Controller{
			&datahub_resources.Controller{
				ObjectMeta: &datahub_resources.ObjectMeta{
					Name:        req.NamespacedName.Name,
					Namespace:   req.NamespacedName.Namespace,
					ClusterName: r.ClusterUID,
				},
				Kind: datahub_resources.Kind_DEPLOYMENT,
			},
		}, nil)
		if err != nil {
			scope.Errorf("Delete controller %s/%s from datahub failed: %s",
				req.NamespacedName.Namespace, req.NamespacedName.Name, err.Error())
		}
	} else if err != nil {
		scope.Errorf("Get Deployment %s/%s failed: %s", req.Namespace, req.Name, err.Error())
		return ctrl.Result{}, nil
	} else {
		alamedaScaler, err := getResource.GetObservingAlamedaScalerOfController(autoscalingv1alpha1.DeploymentController, req.Namespace, req.Name)
		if err != nil && !k8sErrors.IsNotFound(err) {
			scope.Errorf("Get observing AlamedaScaler of Deployment failed: %s", err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: requeueDuration}, nil
		} else if alamedaScaler == nil {
			scope.Warnf("Get observing AlamedaScaler of Deployment %s/%s not found", req.Namespace, req.Name)
		}

		var currentMonitorAlamedaScalerName = ""
		if alamedaScaler != nil {
			if err := controllerutil.TriggerAlamedaScaler(updateResource, alamedaScaler); err != nil {
				scope.Errorf("Trigger current monitoring AlamedaScaler to update falied: %s", err.Error())
				return ctrl.Result{Requeue: true, RequeueAfter: requeueDuration}, nil
			}
			currentMonitorAlamedaScalerName = alamedaScaler.Name
		}

		lastMonitorAlamedaScalerName := controllerutil.GetLastMonitorAlamedaScaler(&deployment)
		// Do not trigger the update process twice if last and current AlamedaScaler are the same
		if lastMonitorAlamedaScalerName != "" && currentMonitorAlamedaScalerName != lastMonitorAlamedaScalerName {
			lastMonitorAlamedaScaler, err := getResource.GetAlamedaScaler(req.Namespace, lastMonitorAlamedaScalerName)
			if err != nil && !k8sErrors.IsNotFound(err) {
				scope.Errorf("Get last monitoring AlamedaScaler falied: %s", err.Error())
				return ctrl.Result{Requeue: true, RequeueAfter: requeueDuration}, nil
			} else if k8sErrors.IsNotFound(err) {
				return ctrl.Result{Requeue: false}, nil
			}
			if lastMonitorAlamedaScaler != nil {
				err := controllerutil.TriggerAlamedaScaler(updateResource, lastMonitorAlamedaScaler)
				if err != nil {
					scope.Errorf("Trigger last monitoring AlamedaScaler to update falied: %s", err.Error())
					return ctrl.Result{Requeue: true, RequeueAfter: requeueDuration}, nil
				}
			}
		}

		controllerutil.SetLastMonitorAlamedaScaler(&deployment, currentMonitorAlamedaScalerName)
		err = updateResource.UpdateResource(&deployment)
		if err != nil {
			scope.Errorf("Update Deployment falied: %s", err.Error())
			return ctrl.Result{Requeue: true, RequeueAfter: requeueDuration}, nil
		}
	}

	return ctrl.Result{}, nil
}

func (r *DeploymentReconciler) getMonitoringAlamedaScaler(namespace, name string) (*autoscalingv1alpha1.AlamedaScaler, error) {

	listResource := utilsresource.NewListResources(r.Client)
	alamedaScalers, err := listResource.ListNamespaceAlamedaScaler(namespace)
	if err != nil {
		return nil, errors.Wrap(err, "list AlamedaScalers failed")
	}

	for _, alamedaScaler := range alamedaScalers {
		for _, deployment := range alamedaScaler.Status.AlamedaController.Deployments {
			if deployment.Namespace == namespace && deployment.Name == name {
				return &alamedaScaler, nil
			}
		}
	}

	return nil, nil
}

func (r *DeploymentReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appsv1.Deployment{}).
		Complete(r)
}
