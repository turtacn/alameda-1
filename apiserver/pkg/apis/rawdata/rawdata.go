package rawdata

import (
	APIServerConfig "github.com/containers-ai/alameda/apiserver/pkg/config"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
)

var (
	scope = Log.RegisterScope("apiserver", "apiserver log", 0)
)

type ServiceRawdata struct {
	Config *APIServerConfig.Config
}

func NewServiceRawdata(cfg *APIServerConfig.Config) *ServiceRawdata {
	service := ServiceRawdata{}
	service.Config = cfg
	return &service
}
