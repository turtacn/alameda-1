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
	"context"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"

	datahub_node "github.com/containers-ai/alameda/operator/datahub/client/node"
	nodeinfo "github.com/containers-ai/alameda/operator/pkg/nodeinfo"

	datahubv1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	datahub_resources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"

	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// NodeReconciler reconciles a Node object
type NodeReconciler struct {
	client.Client
	Scheme *runtime.Scheme

	conn            *grpc.ClientConn
	datahubClient   datahubv1alpha1.DatahubServiceClient
	DatahubNodeRepo datahub_node.AlamedaNodeRepository

	Cloudprovider string
	RegionName    string
	ClusterUID    string
}

// Reconcile reads that state of the cluster for a Node object and makes changes based on the state read
// and what is in the Node.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=nodes/status,verbs=get;update;patch
func (r *NodeReconciler) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	requeueInterval := 3 * time.Second
	instance := &corev1.Node{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)

	nodeIsDeleted := false
	if err != nil && k8sErrors.IsNotFound(err) {
		nodeIsDeleted = true
		instance.Namespace = request.Namespace
		instance.Name = request.Name
	} else if err != nil {
		scope.Error(err.Error())
	}

	nodes := make([]*corev1.Node, 1)
	nodes[0] = instance
	switch nodeIsDeleted {
	case false:
		if err := r.createNodesToDatahub(nodes); err != nil {
			scope.Errorf("Create node to Datahub failed failed: %s", err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: requeueInterval}, nil
		}
	case true:
		if err := r.deleteNodesFromDatahub(nodes); err != nil {
			scope.Errorf("Delete nodes from Datahub failed: %s", err.Error())
			return reconcile.Result{Requeue: true, RequeueAfter: requeueInterval}, nil
		}
	}
	return reconcile.Result{}, nil
}

func (r *NodeReconciler) createNodesToDatahub(nodes []*corev1.Node) error {

	nodeInfos, err := r.createNodeInfos(nodes)
	if err != nil {
		return errors.Wrap(err, "create nodeInfos failed")
	}

	datahubNodes := make([]*datahub_resources.Node, len(nodes))
	for i, nodeInfo := range nodeInfos {
		n := nodeInfo.DatahubNode(r.ClusterUID)
		datahubNodes[i] = &n
	}

	return r.DatahubNodeRepo.CreateNodes(datahubNodes)
}

func (r *NodeReconciler) deleteNodesFromDatahub(nodes []*corev1.Node) error {

	nodeInfos, err := r.createNodeInfos(nodes)
	if err != nil {
		return errors.Wrap(err, "create nodeInfos failed")
	}

	datahubNodes := make([]*datahub_resources.Node, len(nodes))
	for i, nodeInfo := range nodeInfos {
		n := nodeInfo.DatahubNode(r.ClusterUID)
		datahubNodes[i] = &n
	}

	return r.DatahubNodeRepo.DeleteNodes(datahubNodes)
}

func (r *NodeReconciler) createNodeInfos(nodes []*corev1.Node) ([]*nodeinfo.NodeInfo, error) {
	nodeInfos := make([]*nodeinfo.NodeInfo, len(nodes))
	for i, node := range nodes {
		n, err := r.createNodeInfo(node)
		if err != nil {
			return nodeInfos, errors.Wrap(err, "create nodeInfos failed")
		}
		nodeInfos[i] = n
	}
	return nodeInfos, nil
}

func (r *NodeReconciler) createNodeInfo(node *corev1.Node) (*nodeinfo.NodeInfo, error) {
	n, err := nodeinfo.NewNodeInfo(*node)
	if err != nil {
		return nil, errors.Wrap(err, "new NodeInfo failed")
	}
	r.setNodeInfoDefault(&n)
	return &n, nil
}

func (r *NodeReconciler) setNodeInfoDefault(nodeInfo *nodeinfo.NodeInfo) {

	if nodeInfo.Provider == "" {
		nodeInfo.Provider = r.Cloudprovider
	}
	if nodeInfo.Region == "" {
		nodeInfo.Region = r.RegionName
	}
}

func (r *NodeReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1.Node{}).
		Complete(r)
}
