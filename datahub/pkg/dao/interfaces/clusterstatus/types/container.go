package types

import (
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	"github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

// ContainerOperation provides container measurement operations
type ContainerDAO interface {
	AddPods([]*resources.Pod) error
	DeletePods([]*resources.Pod) error
	ListAlamedaPods(string, string, resources.Kind, *common.TimeRange) ([]*resources.Pod, error)
}
