package v1alpha1

import (
	DatahubConfig "github.com/containers-ai/alameda/datahub/pkg/config"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

var (
	scope = Log.RegisterScope("datahub", "datahub(v1alpha1) log", 0)
)

type ServiceV1alpha1 struct {
	Config    *DatahubConfig.Config
	K8SClient client.Client
}

func NewService(cfg *DatahubConfig.Config, k8sClient client.Client) *ServiceV1alpha1 {
	service := ServiceV1alpha1{}
	service.Config = cfg
	service.K8SClient = k8sClient
	return &service
}
