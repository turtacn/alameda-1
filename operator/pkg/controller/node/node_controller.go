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

package node

import (
	"context"
	"time"

	"github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"github.com/pkg/errors"
	"google.golang.org/grpc"

	datahub_node "github.com/containers-ai/alameda/operator/datahub/client/node"
	"github.com/containers-ai/alameda/operator/pkg/controller/firstsync"
	nodeinfo "github.com/containers-ai/alameda/operator/pkg/nodeinfo"
	datahubutils "github.com/containers-ai/alameda/operator/pkg/utils/datahub"
	"github.com/containers-ai/alameda/pkg/provider"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	datahubv1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"

	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
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
	firstSynchronizer firstsync.FirstSynchronizer
	scope             = logUtil.RegisterScope("node_controller", "node controller log", 0)
	requeueInterval   = 3 * time.Second
	grpcDefaultRetry  = uint(3)
)

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Node Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// GetFirstSynchronizer returns reconciler as the FirstSynchronizer
func GetFirstSynchronizer() firstsync.FirstSynchronizer {
	return firstSynchronizer
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	cloudprovider := ""
	if provider.OnGCE() {
		cloudprovider = provider.GCP
	} else if provider.OnEC2() {
		cloudprovider = provider.AWS
	}

	regionName := ""
	switch cloudprovider {
	case provider.AWS:
		regionName = provider.AWSRegionMap[provider.GetEC2Region()]
	}

	conn, _ := grpc.Dial(datahubutils.GetDatahubAddress(), grpc.WithInsecure(), grpc.WithUnaryInterceptor(grpc_retry.UnaryClientInterceptor(grpc_retry.WithMax(grpcDefaultRetry))))
	datahubNodeRepo := datahub_node.NewAlamedaNodeRepository(conn)

	r := ReconcileNode{
		Client: mgr.GetClient(),
		scheme: mgr.GetScheme(),

		datahubNodeRepo: *datahubNodeRepo,

		cloudprovider: cloudprovider,
		regionName:    regionName,
	}
	firstSynchronizer = &r
	return &r
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("node-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Node
	err = c.Watch(&source.Kind{Type: &corev1.Node{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileNode{}

// ReconcileNode reconciles a Node object
type ReconcileNode struct {
	client.Client
	scheme *runtime.Scheme

	conn            *grpc.ClientConn
	datahubClient   datahubv1alpha1.DatahubServiceClient
	datahubNodeRepo datahub_node.AlamedaNodeRepository

	cloudprovider string
	regionName    string
}

// FirstSync synchronizes k8s nodes with Datahub
func (r *ReconcileNode) FirstSync() error {

	// Create existing nodes to Datahub
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	nodeList := corev1.NodeList{}
	if err := r.Client.List(ctx, nil, &nodeList); err != nil {
		return errors.Errorf("get node list failed: %s", err.Error())
	}
	nodes := make([]*corev1.Node, len(nodeList.Items))
	for i, node := range nodeList.Items {
		tmpNode := node
		nodes[i] = &tmpNode
	}
	if err := r.createNodesToDatahub(nodes); err != nil {
		return errors.Wrap(err, "create nodes to Datahub failed")
	}

	// Clean up unexisting nodes from Datahub
	existingNodeMap := make(map[string]bool)
	for _, node := range nodeList.Items {
		existingNodeMap[node.Name] = true
	}
	nodesFromDatahub, err := r.datahubNodeRepo.ListAlamedaNodes()
	if err != nil {
		return errors.Wrap(err, "list nodes from Datahub failed")
	}
	nodesNeedDeleting := make([]*datahubv1alpha1.Node, 0)
	for _, n := range nodesFromDatahub {
		if _, exist := existingNodeMap[n.Name]; exist {
			continue
		}
		nodeInfo, err := r.createNodeInfo(&corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: n.Name}})
		if err != nil {
			return errors.Wrap(err, "create nodeInfo failed")
		}
		datahubNode := nodeInfo.DatahubNode()
		nodesNeedDeleting = append(nodesNeedDeleting, &datahubNode)
	}
	err = r.datahubNodeRepo.DeleteAlamedaNodes(nodesNeedDeleting)
	if err != nil {
		return errors.Wrap(err, "delete nodes from Datahub failed")
	}

	return nil
}

// Reconcile reads that state of the cluster for a Node object and makes changes based on the state read
// and what is in the Node.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// +kubebuilder:rbac:groups=core,resources=nodes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=nodes/status,verbs=get;update;patch
func (r *ReconcileNode) Reconcile(request reconcile.Request) (reconcile.Result, error) {

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

func (r *ReconcileNode) createNodesToDatahub(nodes []*corev1.Node) error {

	nodeInfos, err := r.createNodeInfos(nodes)
	if err != nil {
		return errors.Wrap(err, "create nodeInfos failed")
	}

	datahubNodes := make([]*datahubv1alpha1.Node, len(nodes))
	for i, nodeInfo := range nodeInfos {
		n := nodeInfo.DatahubNode()
		datahubNodes[i] = &n
	}

	return r.datahubNodeRepo.CreateAlamedaNode(datahubNodes)
}

func (r *ReconcileNode) deleteNodesFromDatahub(nodes []*corev1.Node) error {

	nodeInfos, err := r.createNodeInfos(nodes)
	if err != nil {
		return errors.Wrap(err, "create nodeInfos failed")
	}

	datahubNodes := make([]*datahubv1alpha1.Node, len(nodes))
	for i, nodeInfo := range nodeInfos {
		n := nodeInfo.DatahubNode()
		datahubNodes[i] = &n
	}

	return r.datahubNodeRepo.DeleteAlamedaNodes(datahubNodes)
}

func (r *ReconcileNode) createNodeInfos(nodes []*corev1.Node) ([]*nodeinfo.NodeInfo, error) {
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

func (r *ReconcileNode) createNodeInfo(node *corev1.Node) (*nodeinfo.NodeInfo, error) {
	n, err := nodeinfo.NewNodeInfo(*node)
	if err != nil {
		return nil, errors.Wrap(err, "new NodeInfo failed")
	}
	r.setNodeInfoDefault(&n)
	return &n, nil
}

func (r *ReconcileNode) setNodeInfoDefault(nodeInfo *nodeinfo.NodeInfo) {

	if nodeInfo.Provider == "" {
		nodeInfo.Provider = r.cloudprovider
	}
	if nodeInfo.Region == "" {
		nodeInfo.Region = r.regionName
	}
}
