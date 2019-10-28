package clusterstatus

import (
	ApiCommon "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/common"
	ApiResources "github.com/containers-ai/api/alameda_api/v1alpha1/datahub/resources"
)

// ContainerOperation provides container measurement operations
type ContainerOperation interface {
	AddPods([]*ApiResources.Pod) error
	DeletePods([]*ApiResources.Pod) error
	ListAlamedaPods(string, string, ApiResources.Kind, *ApiCommon.TimeRange) ([]*ApiResources.Pod, error)
}
