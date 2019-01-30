package resources

import (
	"context"
	"fmt"
	"strings"

	logUtil "github.com/containers-ai/alameda/operator/pkg/utils/log"
	openshift_appsapi_v1 "github.com/openshift/api/apps/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	resourceLinkScope = logUtil.RegisterScope("resourcelink", "resource link", 0)
)

// GetResourceLinkForPod returns resource link for pod
func GetResourceLinkForPod(client client.Client, pod *corev1.Pod) string {
	controllerOwnerRef := getControllerOwnerRef(pod.OwnerReferences)
	podName := pod.GetName()
	podNS := pod.GetNamespace()
	if controllerOwnerRef == nil {
		return fmt.Sprintf("/namespaces/%s/pods/%s", podNS, podName)
	} else {
		return fmt.Sprintf("/namespaces/%s%s/pods/%s", podNS, getControlleHierarchy(client, *controllerOwnerRef, podNS, ""), podName)
	}
}

func getControlleHierarchy(client client.Client, curOwnerRef metav1.OwnerReference, namespace, curHierarchyLink string) string {
	orKind := curOwnerRef.Kind
	orName := curOwnerRef.Name
	resultStr := ""
	if strings.ToLower(orKind) == "deployment" {
		deploymentFound := &appsv1.Deployment{}
		err := client.Get(context.TODO(), types.NamespacedName{
			Namespace: namespace,
			Name:      orName,
		}, deploymentFound)
		if err != nil {
			resourceLinkScope.Error(err.Error())
		} else {
			controllerOwnerRef := getControllerOwnerRef(deploymentFound.OwnerReferences)
			resultStr = fmt.Sprintf("/deployments/%s", deploymentFound.GetName())
			if controllerOwnerRef == nil {
				return resultStr
			} else {
				return getControlleHierarchy(client, *controllerOwnerRef, namespace, resultStr) + resultStr
			}
		}
	} else if strings.ToLower(orKind) == "replicaset" {
		replicasetFound := &appsv1.ReplicaSet{}
		err := client.Get(context.TODO(), types.NamespacedName{
			Namespace: namespace,
			Name:      orName,
		}, replicasetFound)
		if err != nil {
			resourceLinkScope.Error(err.Error())
		} else {
			controllerOwnerRef := getControllerOwnerRef(replicasetFound.OwnerReferences)
			resultStr = fmt.Sprintf("/replicasets/%s", replicasetFound.GetName())
			if controllerOwnerRef == nil {
				return resultStr
			} else {
				return getControlleHierarchy(client, *controllerOwnerRef, namespace, resultStr) + resultStr
			}
		}
	} else if strings.ToLower(orKind) == "deploymentconfig" {
		deploymentConfigFound := &openshift_appsapi_v1.DeploymentConfig{}
		err := client.Get(context.TODO(), types.NamespacedName{
			Namespace: namespace,
			Name:      orName,
		}, deploymentConfigFound)
		if err != nil {
			resourceLinkScope.Error(err.Error())
		} else {
			controllerOwnerRef := getControllerOwnerRef(deploymentConfigFound.OwnerReferences)
			resultStr = fmt.Sprintf("/deploymentconfigs/%s", deploymentConfigFound.GetName())
			if controllerOwnerRef == nil {
				return resultStr
			} else {
				return getControlleHierarchy(client, *controllerOwnerRef, namespace, resultStr) + resultStr
			}
		}
	} else if strings.ToLower(orKind) == "replicationcontroller" {
		replicationControllerFound := &corev1.ReplicationController{}
		err := client.Get(context.TODO(), types.NamespacedName{
			Namespace: namespace,
			Name:      orName,
		}, replicationControllerFound)
		if err != nil {
			resourceLinkScope.Error(err.Error())
		} else {
			controllerOwnerRef := getControllerOwnerRef(replicationControllerFound.OwnerReferences)
			resultStr = fmt.Sprintf("/replicationcontrollers/%s", replicationControllerFound.GetName())
			if controllerOwnerRef == nil {
				return resultStr
			} else {
				return getControlleHierarchy(client, *controllerOwnerRef, namespace, resultStr) + resultStr
			}
		}
	}
	return resultStr
}

func getControllerOwnerRef(ownerRefs []metav1.OwnerReference) *metav1.OwnerReference {
	for _, or := range ownerRefs {
		if *or.Controller {
			return &or
		}
	}
	return nil
}
