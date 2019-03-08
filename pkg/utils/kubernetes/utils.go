package kubernetes

import (
	"github.com/containers-ai/alameda/pkg/consts"
	logUtil "github.com/containers-ai/alameda/pkg/utils/log"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/rest"
)

var scope = logUtil.RegisterScope("kubernetes_utils", "Kubernetes utils.", 0)

func IsOKDCluster() (bool, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return false, err
	}
	discoveryClient, err := discovery.NewDiscoveryClientForConfig(config)
	if err != nil {
		return false, err
	}
	apiResourceLists, err := discoveryClient.ServerResources()
	if err != nil {
		return false, err
	}

	for _, apiResourceList := range apiResourceLists {
		for _, resource := range apiResourceList.APIResources {
			if resource.Kind == consts.K8S_KIND_DEPLOYMENTCONFIG {
				return true, nil
			}
		}
	}
	return false, nil
}
