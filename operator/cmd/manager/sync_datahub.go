package main

import (
	"fmt"
	"time"

	datahub_node "github.com/containers-ai/alameda/operator/datahub/client/node"
	datahub_pod "github.com/containers-ai/alameda/operator/datahub/client/pod"
	"github.com/containers-ai/alameda/operator/pkg/utils/resources"
	datahub_v1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func registerNodes(client client.Client) {
	time.Sleep(3 * time.Second)
	listResources := resources.NewListResources(client)
	nodes, err := listResources.ListAllNodes()
	if err != nil {
		scope.Errorf("register nodes to Datahub failed: %s", err.Error())
		return
	}
	scope.Infof(fmt.Sprintf("%v nodes found in cluster.", len(nodes)))
	datahubNodeRepo := datahub_node.NewAlamedaNodeRepository()
	datahubNodeRepo.CreateAlamedaNode(nodes)
}

func syncAlamedaPodsWithDatahub(client client.Client) {
	time.Sleep(3 * time.Second)
	datahubPodRepo := datahub_pod.NewPodRepository()
	alamedaPods, _ := datahubPodRepo.ListAlamedaPods()
	podsNeedToRm := []*datahub_v1alpha1.Pod{}
	getResource := resources.NewGetResource(client)
	for _, alamedaPod := range alamedaPods {
		namespacedName := alamedaPod.GetNamespacedName()
		_, err := getResource.GetPod(namespacedName.GetNamespace(), namespacedName.GetName())
		if err != nil && k8sErrors.IsNotFound(err) {
			podsNeedToRm = append(podsNeedToRm, alamedaPod)
			continue
		}
		alamedaScaler := alamedaPod.GetAlamedaScaler()
		_, err = getResource.GetAlamedaScaler(alamedaScaler.GetNamespace(), alamedaScaler.GetName())
		if err != nil && k8sErrors.IsNotFound(err) {
			podsNeedToRm = append(podsNeedToRm, alamedaPod)
			continue
		}
	}
	if len(podsNeedToRm) > 0 {
		err := datahubPodRepo.DeletePods(podsNeedToRm)
		if err != nil {
			scope.Errorf("Remove pods not existed from datahub failed. %s", err.Error())
		}
	}
}
