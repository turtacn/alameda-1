package rawdata

import (
	DatapipeConfig "github.com/containers-ai/alameda/datapipe/pkg/config"
	Log "github.com/containers-ai/alameda/pkg/utils/log"
)

var (
	scope = Log.RegisterScope("datapipe", "datapipe log", 0)
)

type ServiceRawdata struct {
	Config *DatapipeConfig.Config
}

func NewServiceRawdata(cfg *DatapipeConfig.Config) *ServiceRawdata {
	service := ServiceRawdata{}
	service.Config = cfg
	return &service
}
