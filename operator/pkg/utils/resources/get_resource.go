package resources

import (
	"context"
	"fmt"

	autuscaling "github.com/containers-ai/alameda/operator/pkg/apis/autoscaling/v1alpha1"
	logUtil "github.com/containers-ai/alameda/operator/pkg/utils/log"
	appsapi_v1 "github.com/openshift/api/apps/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	getResourcesScope = logUtil.RegisterScope("getresources", "Get resource", 0)
)

// GetResource define resource list functions
type GetResource struct {
	client.Client
}

// NewGetResource return GetResource instance
func NewGetResource(client client.Client) *GetResource {
	return &GetResource{
		client,
	}
}

// GetPod returns pod
func (getResource *GetResource) GetPod(namespace, name string) (*corev1.Pod, error) {
	pod := &corev1.Pod{}
	err := getResource.getResource(pod, namespace, name)
	return pod, err
}

// GetDeploymentConfig returns deploymentconfig
func (getResource *GetResource) GetDeploymentConfig(namespace, name string) (*appsapi_v1.DeploymentConfig, error) {
	deploymentConfig := &appsapi_v1.DeploymentConfig{}
	err := getResource.getResource(deploymentConfig, namespace, name)
	return deploymentConfig, err
}

// GetDeployment returns deployment
func (getResource *GetResource) GetDeployment(namespace, name string) (*appsv1.Deployment, error) {
	deployment := &appsv1.Deployment{}
	err := getResource.getResource(deployment, namespace, name)
	return deployment, err
}

// GetAlamedaScaler return alamedascaler
func (getResource *GetResource) GetAlamedaScaler(namespace, name string) (*autuscaling.AlamedaScaler, error) {
	alamedaScaler := &autuscaling.AlamedaScaler{}
	err := getResource.getResource(alamedaScaler, namespace, name)
	return alamedaScaler, err
}

// GetAlamedaRecommendation return AlamedaRecommendation
func (getResource *GetResource) GetAlamedaRecommendation(namespace, name string) (*autuscaling.AlamedaRecommendation, error) {
	alamedaRecommendation := &autuscaling.AlamedaRecommendation{}
	err := getResource.getResource(alamedaRecommendation, namespace, name)
	return alamedaRecommendation, err
}

func (getResource *GetResource) getResource(resource runtime.Object, namespace, name string) error {
	if namespace == "" || name == "" {
		return fmt.Errorf("Namespace: %s or name: %s is empty", namespace, name)
	}
	if err := getResource.Get(context.TODO(),
		types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		},
		resource); err != nil {
		getResourcesScope.Error(err.Error())
		return err
	}
	return nil
}
