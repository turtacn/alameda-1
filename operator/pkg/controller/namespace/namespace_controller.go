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

package namespace

import (
	"time"

	datahub_namespace "github.com/containers-ai/alameda/operator/datahub/client/namespace"
	datahubutils "github.com/containers-ai/alameda/operator/pkg/utils/datahub"
	k8s_utils "github.com/containers-ai/alameda/pkg/utils/kubernetes"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
	grpc_retry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/pkg/errors"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
	corev1 "k8s.io/api/core/v1"
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
	scope = logUtil.RegisterScope(
		"namespace_controller", "namespace controller log", 0)
	cachedFirstSynced = false
	requeueDuration   = 1 * time.Second
	grpcDefaultRetry  = uint(3)
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Namespace Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	conn, _ := grpc.Dial(datahubutils.GetDatahubAddress(),
		grpc.WithInsecure(), grpc.WithUnaryInterceptor(
			grpc_retry.UnaryClientInterceptor(grpc_retry.WithMax(grpcDefaultRetry))))

	k8sClient, err := client.New(mgr.GetConfig(), client.Options{})
	if err != nil {
		panic(errors.Wrap(err, "new kuberenetes client failed").Error())
	}
	clusterUID, err := k8s_utils.GetClusterUID(k8sClient)
	if err != nil || clusterUID == "" {
		panic("cannot get cluster uid")
	}

	datahubNamespaceRepo := datahub_namespace.NewNamespaceRepository(conn, clusterUID)

	return &ReconcileNamespace{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),

		clusterUID: clusterUID,

		datahubNamespaceRepo: datahubNamespaceRepo,
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("namespace-controller",
		mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Namespace
	err = c.Watch(&source.Kind{Type: &corev1.Namespace{}},
		&handler.EnqueueRequestForObject{})
	if err != nil {
		scope.Error(err.Error())
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileNamespace{}

// ReconcileNamespace reconciles a Namespace object
type ReconcileNamespace struct {
	client.Client
	scheme *runtime.Scheme

	clusterUID string

	datahubNamespaceRepo *datahub_namespace.NamespaceRepository
}

func (r *ReconcileNamespace) Reconcile(
	request reconcile.Request) (reconcile.Result, error) {

	if !cachedFirstSynced {
		time.Sleep(5 * time.Second)
	}
	cachedFirstSynced = true
	namespace := corev1.Namespace{}
	err := r.Get(context.Background(), request.NamespacedName, &namespace)
	if err != nil && k8sErrors.IsNotFound(err) {
		err = r.datahubNamespaceRepo.DeleteNamespaces(
			[]*datahub_resources.Namespace{
				&datahub_resources.Namespace{
					ObjectMeta: &datahub_resources.ObjectMeta{
						Name:        request.NamespacedName.Name,
						ClusterName: r.clusterUID,
					},
				},
			})
		if err != nil {
			scope.Errorf("Delete namespace %s from datahub failed: %s",
				request.NamespacedName.Name, err.Error())
		}
	} else if err == nil {
		err = r.datahubNamespaceRepo.CreateNamespaces(
			[]*datahub_resources.Namespace{
				&datahub_resources.Namespace{
					ObjectMeta: &datahub_resources.ObjectMeta{
						Name:        request.NamespacedName.Name,
						ClusterName: r.clusterUID,
					},
				},
			})
		if err != nil {
			scope.Errorf("create namespace %s from datahub failed: %s",
				request.NamespacedName.Name, err.Error())
		}
	}
	return reconcile.Result{}, nil
}
