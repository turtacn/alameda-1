package resources

import (
	"context"
	"fmt"
	"strings"

	logUtil "github.com/containers-ai/alameda/operator/pkg/utils/log"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ListPods struct {
	client client.Client
}

func NewListPods(client client.Client) *ListPods {
	return &ListPods{
		client: client,
	}
}

func (listpods *ListPods) ListPods(namespace, name, kind string) []corev1.Pod {
	podList := []corev1.Pod{}
	deploymentFound := &appsv1.Deployment{}
	if strings.ToLower(kind) == "deployment" {
		err := listpods.client.Get(context.TODO(), types.NamespacedName{
			Namespace: namespace,
			Name:      name,
		}, deploymentFound)
		if err != nil {
			logUtil.GetLogger().Error(err, "Get pods from deployment failed.")
			return podList
		} else {
			return listpods.getPodsFromDeployment(deploymentFound)
		}
	}
	return podList
}

func (listpods *ListPods) getPodsFromDeployment(deployment *appsv1.Deployment) []corev1.Pod {
	podList := []corev1.Pod{}
	pods := &corev1.PodList{}
	name := deployment.GetName()
	ns := deployment.GetNamespace()
	if deployment.Spec.Selector == nil {
		logUtil.GetLogger().Info(fmt.Sprintf("List pods of alameda deployment %s/%s failed due to no matched labels found.", ns, name))
		return podList
	}
	labels := deployment.Spec.Selector.MatchLabels

	err := listpods.client.List(context.TODO(),
		client.InNamespace(ns).
			MatchingLabels(labels),
		pods)
	if err != nil {
		logUtil.GetLogger().Info(fmt.Sprintf("List pods of alameda deployment %s/%s failed.", ns, name))
	} else {
		var deploymentName string
		for _, pod := range pods.Items {
			for _, ownerReference := range pod.ObjectMeta.GetOwnerReferences() {

				if ownerReference.Kind == "ReplicaSet" {
					replicaSetName := ownerReference.Name
					deploymentName = replicaSetName[0:strings.LastIndex(replicaSetName, "-")]
				}
				break
			}
			if deploymentName == name {
				podList = append(podList, pod)
			}
		}
	}
	logUtil.GetLogger().Info(fmt.Sprintf("%d pods founded in alameda deployment %s/%s.", len(podList), ns, name))
	return podList
}
