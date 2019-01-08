package resources

import (
	"context"
	"fmt"

	logUtil "github.com/containers-ai/alameda/operator/pkg/utils/log"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	listResourcesScope = logUtil.RegisterScope("listresources", "List resources", 0)
)

// ListResources define resource list functions
type ListResources struct {
	client.Client
}

// NewListResources return ListResources instance
func NewListResources(client client.Client) *ListResources {
	return &ListResources{
		client,
	}
}

// ListAllNodes return all nodes in cluster
func (ListResources *ListResources) ListAllNodes() *corev1.NodeList {
	listResourcesScope.Infof(fmt.Sprintf("%s,%s", labels.Everything().String(), fields.Everything().String()))
	nodeList := &corev1.NodeList{}
	if err := ListResources.List(context.TODO(),
		&client.ListOptions{},
		nodeList); err != nil {
		listResourcesScope.Error(err.Error())
	}
	return nodeList
}
