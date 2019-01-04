package clusterstatus

import (
	datahub_api "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

// ContainerOperation provides container measurement operations
type ContainerOperation interface {
	AddPods([]*datahub_api.Pod) error
	UpdatePods([]*datahub_api.Pod) error
	ListAlamedaPods() ([]*datahub_api.Pod, error)
}
