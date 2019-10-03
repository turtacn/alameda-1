package requests

import (
	DatahubV1alpha1 "github.com/containers-ai/api/alameda_api/v1alpha1/datahub"
)

type ListPodMetricsRequestExtended struct {
	DatahubV1alpha1.ListPodMetricsRequest
}

func (r *ListPodMetricsRequestExtended) Validate() error {
	return nil
}

type ListNodeMetricsRequestExtended struct {
	DatahubV1alpha1.ListNodeMetricsRequest
}

func (r *ListNodeMetricsRequestExtended) Validate() error {
	return nil
}
