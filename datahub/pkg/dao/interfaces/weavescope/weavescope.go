package weavescope

import (
	"github.com/containers-ai/alameda/internal/pkg/weavescope"
	ApiWeavescope "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/weavescope"
)

// Container Implements ContainerOperation interface
type WeaveScope struct {
	WeaveScopeConfig *weavescope.Config
}

func (w *WeaveScope) ListWeaveScopeHosts(in *ApiWeavescope.ListWeaveScopeHostsRequest) (string, error) {
	weaveScopeRepository := weavescope.NewClient(w.WeaveScopeConfig)
	return weaveScopeRepository.ListWeaveScopeHosts(in)
}

func (w *WeaveScope) GetWeaveScopeHostDetails(in *ApiWeavescope.ListWeaveScopeHostsRequest) (string, error) {
	weaveScopeRepository := weavescope.NewClient(w.WeaveScopeConfig)
	return weaveScopeRepository.GetWeaveScopeHostDetails(in)
}

func (w *WeaveScope) ListWeaveScopePods(in *ApiWeavescope.ListWeaveScopePodsRequest) (string, error) {
	weaveScopeRepository := weavescope.NewClient(w.WeaveScopeConfig)
	return weaveScopeRepository.ListWeaveScopePods(in)
}

func (w *WeaveScope) GetWeaveScopePodDetails(in *ApiWeavescope.ListWeaveScopePodsRequest) (string, error) {
	weaveScopeRepository := weavescope.NewClient(w.WeaveScopeConfig)
	return weaveScopeRepository.GetWeaveScopePodDetails(in)
}

func (w *WeaveScope) ListWeaveScopeContainers(in *ApiWeavescope.ListWeaveScopeContainersRequest) (string, error) {
	weaveScopeRepository := weavescope.NewClient(w.WeaveScopeConfig)
	return weaveScopeRepository.ListWeaveScopeContainers(in)
}

func (w *WeaveScope) ListWeaveScopeContainersByHostname(in *ApiWeavescope.ListWeaveScopeContainersRequest) (string, error) {
	weaveScopeRepository := weavescope.NewClient(w.WeaveScopeConfig)
	return weaveScopeRepository.ListWeaveScopeContainersByHostname(in)
}

func (w *WeaveScope) ListWeaveScopeContainersByImage(in *ApiWeavescope.ListWeaveScopeContainersRequest) (string, error) {
	weaveScopeRepository := weavescope.NewClient(w.WeaveScopeConfig)
	return weaveScopeRepository.ListWeaveScopeContainersByImage(in)
}

func (w *WeaveScope) GetWeaveScopeContainerDetails(in *ApiWeavescope.ListWeaveScopeContainersRequest) (string, error) {
	weaveScopeRepository := weavescope.NewClient(w.WeaveScopeConfig)
	return weaveScopeRepository.GetWeaveScopeContainerDetails(in)
}
