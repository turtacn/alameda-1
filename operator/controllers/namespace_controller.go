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

	datahub_namespace "github.com/containers-ai/alameda/operator/datahub/client/namespace"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	"golang.org/x/net/context"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	namespaceFirstSynced = false
)

// NamespaceReconciler reconciles a Namespace object
type NamespaceReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	ClusterUID string

	DatahubNamespaceRepo *datahub_namespace.NamespaceRepository
}

func (r *NamespaceReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	if !namespaceFirstSynced {
		time.Sleep(5 * time.Second)
	}
	namespaceFirstSynced = true
	namespace := corev1.Namespace{}
	err := r.Get(context.Background(), req.NamespacedName, &namespace)
	if err != nil && k8sErrors.IsNotFound(err) {
		err = r.DatahubNamespaceRepo.DeleteNamespaces(
			[]*datahub_resources.Namespace{
				&datahub_resources.Namespace{
					ObjectMeta: &datahub_resources.ObjectMeta{
						Name:        req.NamespacedName.Name,
						ClusterName: r.ClusterUID,
					},
				},
			})
		if err != nil {
			scope.Errorf("Delete namespace %s from datahub failed: %s",
				req.NamespacedName.Name, err.Error())
		}
	} else if err == nil {
		err = r.DatahubNamespaceRepo.CreateNamespaces(
			[]*datahub_resources.Namespace{
				&datahub_resources.Namespace{
					ObjectMeta: &datahub_resources.ObjectMeta{
						Name:        req.NamespacedName.Name,
						ClusterName: r.ClusterUID,
					},
				},
			})
		if err != nil {
			scope.Errorf("create namespace %s from datahub failed: %s",
				req.NamespacedName.Name, err.Error())
		}
	}
	return ctrl.Result{}, nil
}

func (r *NamespaceReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Namespace{}).
		Complete(r)
}
